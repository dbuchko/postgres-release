package deploy_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/postgres-release/src/acceptance-tests/testing/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup and restore a deployment", func() {

	var caCertPath string
	var tempDir string
	var pwd string

	JustBeforeEach(func() {
		var err error
		err = deployHelper.Deploy()
		Expect(err).NotTo(HaveOccurred())

		caCertPath, err = helpers.WriteFile(configParams.Bosh.DirectorCACert)
		Expect(err).NotTo(HaveOccurred())

		tempDir, err = helpers.CreateTempDir()
		Expect(err).NotTo(HaveOccurred())
		pwd, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(tempDir)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("BOSH_CLIENT_SECRET", configParams.Bosh.Password)
		os.Setenv("CA_CERT", caCertPath)
	})

	AfterEach(func() {
		err := os.Chdir(pwd)
		Expect(err).NotTo(HaveOccurred())
		os.Remove(caCertPath)
		os.RemoveAll(tempDir)
	})

	Context("BBR is disabled", func() {

		BeforeEach(func() {
			deployHelper.SetOpDefs(nil)
		})

		It("Fails to run pre-backup-checks", func() {
			var err error
			var cmd *exec.Cmd
			By("Running pre-backup-checks")
			cmd = exec.Command("bbr", "deployment", "--target", configParams.Bosh.Target, "--username", configParams.Bosh.Username, "--deployment", deployHelper.GetDeploymentName(), "pre-backup-check")
			err = cmd.Run()
			Expect(err).To(HaveOccurred())
		})

		It("Fails to backup the database", func() {
			var err error
			var cmd *exec.Cmd
			cmd = exec.Command("bbr", "deployment", "--target", configParams.Bosh.Target, "--username", configParams.Bosh.Username, "--deployment", deployHelper.GetDeploymentName(), "backup")
			err = cmd.Run()
			Expect(err).To(HaveOccurred())
		})

	})

	Context("BBR is enabled", func() {

		BeforeEach(func() {
			deployHelper.SetOpDefs(helpers.Define_bbr_ops())
		})

		It("Successfully backup the database", func() {
			var err error
			var cmd *exec.Cmd
			By("Running pre-backup-checks")
			cmd = exec.Command("bbr", "deployment", "--target", configParams.Bosh.Target, "--username", configParams.Bosh.Username, "--deployment", deployHelper.GetDeploymentName(), "pre-backup-check")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred(), "Check the bbr logfile bbr-TIMESTAMP.err.log for why this has failed")
			By("Running backup")
			cmd = exec.Command("bbr", "deployment", "--target", configParams.Bosh.Target, "--username", configParams.Bosh.Username, "--deployment", deployHelper.GetDeploymentName(), "backup")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred(), "Check the bbr logfile bbr-TIMESTAMP.err.log for why this has failed")
			tarBackupFile := fmt.Sprintf("%s/%s*/postgres-0-postgres.tar", tempDir, deployHelper.GetDeploymentName())
			files, err := filepath.Glob(tarBackupFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).NotTo(BeEmpty())
		})
	})
})
