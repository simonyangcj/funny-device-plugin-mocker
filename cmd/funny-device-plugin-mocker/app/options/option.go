package options

import "github.com/spf13/cobra"

const (
	defaultRootPath     = "/mnt/fake/"
	defaultPrefix       = "fdpm_"
	defaultResourceName = "fake.com/device"
)

type Option struct {
	RootPath     string
	Prefix       string
	ResourceName string
}

func (o *Option) BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RootPath, "root-path", "r", defaultRootPath, "root path of this plugin to fetch sub directories")
	cmd.Flags().StringVarP(&o.Prefix, "prefix", "p", defaultPrefix, "root path of this plugin to fetch sub directories")
	cmd.Flags().StringVarP(&o.ResourceName, "resource-name", "n", defaultResourceName, "name of the resource that display in node description")
}

func NewOptions() *Option {
	return &Option{
		RootPath:     defaultRootPath,
		Prefix:       defaultPrefix,
		ResourceName: defaultResourceName,
	}
}
