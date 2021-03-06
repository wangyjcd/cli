package isolated

import (
	"os"
	"path/filepath"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("download-droplet command", func() {
	var (
		helpText func(*Session)
		appName  string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()

		helpText = func(session *Session) {
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("download-droplet - Download an application droplet"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf download-droplet APP_NAME"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("apps, droplets, push, set-droplet"))
		}
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("download-droplet", "APPS", "Download an application droplet"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("download-droplet", "--help")
				helpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("download-droplet")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "download-droplet", appName)
		})
	})

	When("the environment is setup correctly", func() {
		var (
			spaceName string
			orgName   string
			userName  string
		)

		BeforeEach(func() {
			spaceName = helpers.NewSpaceName()
			orgName = helpers.NewOrgName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app has a current droplet", func() {
			var dropletPath string

			BeforeEach(func() {
				helpers.CreateApp(appName)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
				dropletSession := helpers.CF("droplets", appName)
				Eventually(dropletSession).Should(Exit(0))
				regex := regexp.MustCompile(`(.+)\s+\(current\)`)
				matches := regex.FindStringSubmatch(string(dropletSession.Out.Contents()))
				Expect(matches).To(HaveLen(2))
				dropletGUID := matches[1]

				dir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				dropletPath = filepath.Join(dir, "droplet_"+dropletGUID+".tgz")
			})

			AfterEach(func() {
				os.RemoveAll(dropletPath)
			})

			It("downloads the droplet successfully", func() {
				session := helpers.CF("download-droplet", appName)
				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(helpers.SayPath(`Droplet downloaded successfully at %s`, dropletPath))
				Eventually(session).Should(Say("OK"))

				_, err := os.Stat(dropletPath)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("download-droplet", appName)

				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does not have a current droplet", func() {
			BeforeEach(func() {
				helpers.CreateApp(appName)
			})

			It("displays that there is no current droplet and exits 1", func() {
				session := helpers.CF("download-droplet", appName)

				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' does not have a current droplet.", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
