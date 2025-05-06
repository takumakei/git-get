package flags

import (
	"log/slog"
	"os"
	"strings"

	"github.com/goaux/results"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Init(flagset *pflag.FlagSet, prefix string) {
	flagset.SortFlags = false
	flagset.VisitAll(func(flag *pflag.Flag) {
		key := prefix + strings.ToUpper(initReplacer.Replace(flag.Name))
		if v, ok := os.LookupEnv(key); ok {
			results.Must(flag.Value.Set(v))
			flag.DefValue = v
			flag.Changed = false
		}
		flag.Usage += " [$" + key + "]"
	})
}

var initReplacer = strings.NewReplacer("-", "_")

func Log(cmd *cobra.Command) slog.Attr {
	return slog.Group("flags",
		slog.Group("PersistentFlags", attrs(cmd.PersistentFlags())...),
		slog.Group("Flags", attrs(cmd.Flags())...),
	)
}

func attrs(flagset *pflag.FlagSet) []any {
	var list []any
	flagset.VisitAll(func(flag *pflag.Flag) {
		list = append(list, flag.Name, flag.Value.String())
	})
	return list
}
