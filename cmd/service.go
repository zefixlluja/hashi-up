package cmd

import (
	"fmt"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/jsiebens/hashi-up/scripts"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
)

func ManageServiceCommand(action, product string) *cobra.Command {

	var command = &cobra.Command{
		Use:          action,
		Short:        fmt.Sprintf("%s %s systemd service on a server via SSH", strings.Title(action), strings.Title(product)),
		Long:         fmt.Sprintf("%s %s systemd service on a server via SSH", strings.Title(action), strings.Title(product)),
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.RunE = func(command *cobra.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir)
			if err != nil {
				return fmt.Errorf("error received during preparation: %s", err)
			}

			installScript, err := scripts.Open("service.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/run.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info(fmt.Sprintf("%sing %s ...", strings.Title(action), strings.Title(product)))
			sudoPass, err := target.sudoPass()
			if err != nil {
				return fmt.Errorf("error received during execution: %s", err)
			}
			_, err = op.Execute(fmt.Sprintf("cat %s/run.sh | ACTION=%s SERVICE=%s SUDO_PASS=\"%s\" sh -\n", dir, action, product, sudoPass))
			if err != nil {
				return fmt.Errorf("error received during execution: %s", err)
			}

			info("Done.")

			return nil
		}

		return target.execute(callback)
	}

	return command
}
