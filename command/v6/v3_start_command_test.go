package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-start Command", func() {
	var (
		cmd             V3StartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeV3StartActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeV3StartActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = V3StartCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the actor does not return an error", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			fakeActor.StartApplicationReturns(v3action.Warnings{"start-warning-1", "start-warning-2"}, nil)
		})

		It("says that the app was started and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
			appName := fakeActor.StartApplicationArgsForCall(0)
			Expect(appName).To(Equal("some-app-guid"))
		})
	})

	When("the get app call returns a ApplicationNotFoundError", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).ToNot(Say("Starting"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
		})
	})

	When("the start app call returns a ApplicationNotFoundError (someone else deleted app after we fetched summary)", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.StartApplicationReturns(v3action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
		})
	})

	When("the app is already started", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStarted}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).ToNot(Say("Starting"))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("App some-app is already started"))

			Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
		})
	})

	When("the get application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).ToNot(Say("Starting"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))

			Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
		})
	})

	When("the start application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.StartApplicationReturns(v3action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
		})
	})
})
