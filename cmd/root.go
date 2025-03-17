/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"github.com/theapemachine/lookatthatmongo/mongodb/optimizer"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lookatthatmongo",
	Short: "A brief description of your application",
	Long:  rootLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := mongodb.NewConn(cmd.Context(), os.Getenv("MONGO_URI"), "FanAppDev2")
		if err != nil {
			return err
		}
		defer conn.Close(cmd.Context())

		monitor := mongodb.NewMonitor(mongodb.WithConn(conn))
		report := metrics.NewReport(monitor)

		// Collect all metrics
		err = report.Collect(cmd.Context(), "FanAppDev2", func() ([]string, error) {
			return conn.Database("FanAppDev2").ListCollectionNames(cmd.Context(), struct{}{})
		})
		if err != nil {
			return err
		}

		// Generate optimization suggestions
		aiconn := ai.NewConn()
		prompt := ai.NewPrompt(ai.WithReport("before", report))
		suggestion, err := aiconn.Generate(cmd.Context(), prompt)
		if err != nil {
			return err
		}

		// Apply optimizations
		opt := optimizer.NewOptimizer(
			optimizer.WithConnection(conn),
			optimizer.WithMonitor(monitor),
		)

		if err := opt.Apply(cmd.Context(), suggestion.(*ai.OptimizationSuggestion)); err != nil {
			// Attempt rollback on failure
			if rbErr := opt.Rollback(cmd.Context(), suggestion.(*ai.OptimizationSuggestion)); rbErr != nil {
				return fmt.Errorf("optimization failed and rollback failed: %v (rollback: %v)", err, rbErr)
			}
			return fmt.Errorf("optimization failed but rolled back successfully: %v", err)
		}

		// Validate the changes
		result, err := opt.Validate(cmd.Context(), suggestion.(*ai.OptimizationSuggestion))
		if err != nil {
			return fmt.Errorf("validation failed: %v", err)
		}

		fmt.Printf("Optimization applied successfully. Improvement: %.2f%%\n", result.Improvement)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lookatthatmongo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var rootLong = `
A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.
`
