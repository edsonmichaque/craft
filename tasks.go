package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func generateTasks(projectPath string, cfg Config) error {
	ciDir := filepath.Join(projectPath, "scripts", "ci")

	// Create subdirectories for better organization
	dirs := []string{
		"lib",   // Shared libraries
		"tasks", // Individual task scripts
		"utils", // Utility scripts
		"hooks", // Git hooks
		"env",   // Environment configurations
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(ciDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create CI directory %s: %w", dir, err)
		}
	}

	// Generate script files
	scripts := map[string]string{
		// Core library files
		"lib/common.sh":  commonLibTemplate,
		"lib/logger.sh":  loggerLibTemplate,
		"lib/docker.sh":  dockerLibTemplate,
		"lib/git.sh":     gitLibTemplate,
		"lib/version.sh": versionLibTemplate,

		// Task scripts
		"tasks/build.sh":        buildTaskTemplate,
		"tasks/test.sh":         testTaskTemplate,
		"tasks/lint.sh":         lintTaskTemplate,
		"tasks/release.sh":      releaseTaskTemplate,
		"tasks/docker.sh":       dockerTaskTemplate,
		"tasks/proto.sh":        protoTaskTemplate,
		"tasks/dependencies.sh": dependenciesTaskTemplate,
		"tasks/package.sh":      packageTaskTemplate,

		// Utility scripts
		"utils/health-check.sh": healthCheckTemplate,
		"utils/cleanup.sh":      cleanupTemplate,
		"utils/setup-dev.sh":    setupDevTemplate,

		// Main entry points
		"build":       mainBuildTemplate,
		"test":        mainTestTemplate,
		"ci":          mainCITemplate,
		"README.adoc": readmeTemplate,

		"tests/lib/test_helper.bash": testHelperTemplate,
		"tests/lib/common_test.bats": commonTestTemplate,
		"tests/lib/logger_test.bats": loggerTestTemplate,
		"tests/lib/docker_test.bats": dockerTestTemplate,
	}

	for filename, content := range scripts {
		filepath := path.Join(ciDir, filename)

		// Ensure parent directory exists
		if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate CI script %s: %w", filename, err)
		}

		// Make the script executable
		if err := os.Chmod(filepath, 0755); err != nil {
			return fmt.Errorf("failed to make %s executable: %w", filename, err)
		}
	}

	return nil
}

const dbCommandsTemplate = `package commands


import (
	"fmt"
	"github.com/spf13/cobra"
)

// Database Commands
func NewDBCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long:  "Manage database operations including migrations, backups, and maintenance",
	}

	cmd.AddCommand(
		newDBStatusCommand(),
		newDBMigrateCommand(),
		newDBBackupCommand(),
		newDBRestoreCommand(),
		newDBShellCommand(),
	)

	return cmd
}

func newDBStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check database connection status",
		RunE:  dbStatusHandler,
	}
}

func newDBMigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE:  dbMigrateHandler,
	}
}

func newDBBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup database",
		RunE:  dbBackupHandler,
	}
	cmd.Flags().StringP("output", "o", "", "backup file path")
	cmd.Flags().BoolP("compress", "z", true, "compress backup")
	cmd.Flags().StringP("format", "f", "custom", "backup format (plain/custom)")
	
	return cmd
}

func newDBRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore [file]",
		Short: "Restore database from backup",
		Args:  cobra.ExactArgs(1),
		RunE:  dbRestoreHandler,
	}
	cmd.Flags().BoolP("clean", "c", false, "clean database before restore")
	cmd.Flags().BoolP("single-transaction", "s", true, "restore in single transaction")
	return cmd
}

func newDBShellCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Open database shell",
		RunE:  dbShellHandler,
	}
}
`

const cacheCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Cache Commands
func NewCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache operations",
		Long:  "Manage cache operations including data manipulation and monitoring",
	}

	cmd.AddCommand(
		newCacheFlushCommand(),
		newCacheStatsCommand(),
		newCacheGetCommand(),
		newCacheSetCommand(),
		newCacheDeleteCommand(),
		newCacheListCommand(),
	)

	return cmd
}

func newCacheFlushCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "flush",
		Short: "Flush cache",
		RunE:  cacheFlushHandler,
	}
}

func newCacheStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		RunE:  cacheStatsHandler,
	}
}

func newCacheGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get cache value",
		Args:  cobra.ExactArgs(1),
		RunE:  cacheGetHandler,
	}
}

func newCacheSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set cache value",
		Args:  cobra.ExactArgs(2),
		RunE:  cacheSetHandler,
	}
	cmd.Flags().DurationP("ttl", "t", 0, "time to live")
	cmd.Flags().BoolP("nx", "n", false, "set if not exists")
	cmd.Flags().BoolP("xx", "x", false, "set if exists")
	return cmd
}

func newCacheDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [key]",
		Short: "Delete cache key",
		Args:  cobra.ExactArgs(1),
		RunE:  cacheDeleteHandler,
	}
}

func newCacheListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [pattern]",
		Short: "List cache keys",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cacheListHandler,
	}
}
`

const userCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// User Commands
func NewUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
		
		Long:  "Comprehensive user management including accounts, roles, and permissions",
	}

	cmd.AddCommand(
		newUserCreateCommand(),
		newUserListCommand(),
		newUserUpdateCommand(),
		newUserShowCommand(),
		newUserDeleteCommand(),
		newUserDisableCommand(),
		newUserEnableCommand(),
		newUserResetPasswordCommand(),
		newUserSetPasswordCommand(),
		newUserAddGroupCommand(),
		newUserRemoveGroupCommand(),
		newUserListGroupsCommand(),
		newUserSearchCommand(),
		newUserExportCommand(),
		newUserImportCommand(),
		newUserAuditCommand(),
	)

	return cmd
}

func newUserCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new user",
		Args:  cobra.ExactArgs(1),
		RunE:  userCreateHandler,
	}
	cmd.Flags().StringP("email", "e", "", "user email")
	cmd.Flags().StringP("password", "p", "", "user password")
	cmd.Flags().StringP("role", "r", "user", "user role")
	cmd.Flags().StringSliceP("groups", "g", []string{}, "user groups")
	cmd.Flags().StringToStringP("metadata", "m", nil, "user metadata")
	cmd.Flags().Bool("disabled", false, "create user in disabled state")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	return cmd
}

func newUserListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE:  userListHandler,
	}
	cmd.Flags().StringP("role", "r", "", "filter by role")
	cmd.Flags().StringP("group", "g", "", "filter by group")
	cmd.Flags().StringP("status", "s", "", "filter by status (active/disabled)")
	cmd.Flags().StringP("sort", "o", "username", "sort field")
	cmd.Flags().BoolP("desc", "d", false, "sort in descending order")
	cmd.Flags().Int32P("limit", "l", 50, "maximum number of users")
	cmd.Flags().Int32P("offset", "f", 0, "offset for pagination")
	return cmd
}

func newUserUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [username]",
		Short: "Update user details",
		Args:  cobra.ExactArgs(1),
		RunE:  userUpdateHandler,
	}
	cmd.Flags().StringP("email", "e", "", "new email")
	cmd.Flags().StringP("role", "r", "", "new role")
	cmd.Flags().StringSliceP("groups", "g", []string{}, "user groups")
	cmd.Flags().StringToStringP("metadata", "m", nil, "user metadata")
	return cmd
}

func newUserShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show [username]",
		Short: "Show user details",
		Args:  cobra.ExactArgs(1),
		RunE:  userShowHandler,
	}
}

func newUserDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [username]",
		Short: "Delete user",
		Args:  cobra.ExactArgs(1),
		RunE:  userDeleteHandler,
	}
}

func newUserDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable [username]",
		Short: "Disable user account",
		Args:  cobra.ExactArgs(1),
		RunE:  userDisableHandler,
	}
}

func newUserEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable [username]",
		Short: "Enable user account",
		Args:  cobra.ExactArgs(1),
		RunE:  userEnableHandler,
	}
}

func newUserResetPasswordCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset-password [username]",
		Short: "Reset user password",
		Args:  cobra.ExactArgs(1),
		RunE:  userResetPasswordHandler,
	}
}

func newUserSetPasswordCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-password [username]",
		Short: "Set user password",
		Args:  cobra.ExactArgs(1),
		RunE:  userSetPasswordHandler,
	}
}

func newUserAddGroupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-group [username] [group]",
		Short: "Add user to group",
		Args:  cobra.ExactArgs(2),
		RunE:  userAddGroupHandler,
	}
}

func newUserRemoveGroupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-group [username] [group]",
		Short: "Remove user from group",
		Args:  cobra.ExactArgs(2),
		RunE:  userRemoveGroupHandler,
	}
}

func newUserListGroupsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-groups [username]",
		Short: "List user groups",
		Args:  cobra.ExactArgs(1),
		RunE:  userListGroupsHandler,
	}
}

func newUserSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search users",
		RunE:  userSearchHandler,
	}
}

func newUserExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export users to file",
		RunE:  userExportHandler,
	}
}

func newUserImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import [file]",
		Short: "Import users from file",
		Args:  cobra.ExactArgs(1),
		RunE:  userImportHandler,
	}
}

func newUserAuditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "audit [username]",
		Short: "Show user audit log",
		Args:  cobra.ExactArgs(1),
		RunE:  userAuditHandler,
	}
}
`

const queueCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Queue Commands
func NewQueueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Queue operations",
		Long:  "Manage message queues and message processing",
	}

	cmd.AddCommand(
		newQueuePublishCommand(),
		newQueueConsumeCommand(),
		newQueueListCommand(),
		newQueuePurgeCommand(),
		newQueueDeleteCommand(),
		newQueueStatsCommand(),
	)

	return cmd
}

func newQueuePublishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish [queue] [message]",
		Short: "Publish message to queue",
		Args:  cobra.ExactArgs(2),
		RunE:  queuePublishHandler,
	}
	cmd.Flags().StringP("priority", "p", "normal", "message priority (low/normal/high)")
	cmd.Flags().StringToStringP("headers", "H", nil, "message headers")
	cmd.Flags().DurationP("ttl", "t", 0, "message time-to-live")
	return cmd
}

func newQueueConsumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consume [queue]",
		Short: "Consume messages from queue",
		Args:  cobra.ExactArgs(1),
		RunE:  queueConsumeHandler,
	}
	cmd.Flags().Int32P("prefetch", "n", 1, "prefetch count")
	cmd.Flags().BoolP("ack", "a", true, "enable message acknowledgment")
	cmd.Flags().DurationP("timeout", "t", 0, "consume timeout")
	return cmd
}

func newQueueListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List queues",
		RunE:  queueListHandler,
	}
}

func newQueuePurgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "purge [queue]",
		Short: "Purge queue",
		Args:  cobra.ExactArgs(1),
		RunE:  queuePurgeHandler,
	}
}

func newQueueDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [queue]",
		Short: "Delete queue",
		Args:  cobra.ExactArgs(1),
		RunE:  queueDeleteHandler,
	}
}

func newQueueStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats [queue]",
		Short: "Show queue statistics",
		Args:  cobra.MaximumNArgs(1),
		RunE:  queueStatsHandler,
	}
}
`

const notificationCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Notification Commands
func NewNotificationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Notification operations",
		Long:  "Manage notification operations including sending, listing, and configuring notifications",
	}

	cmd.AddCommand(
		newNotificationSendCommand(),
		newNotificationListCommand(),
		newNotificationConfigCommand(),
	)

	return cmd
}

func newNotificationSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [recipient] [message]",
		Short: "Send a notification",
		Args:  cobra.ExactArgs(2),
		RunE:  notificationSendHandler,
	}
	cmd.Flags().StringP("type", "t", "email", "notification type (email/sms/push)")
	cmd.Flags().StringToStringP("metadata", "m", nil, "additional metadata")
	return cmd
}

func newNotificationListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List notifications",
		RunE:  notificationListHandler,
	}
}

func newNotificationConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure notification settings",
		RunE:  notificationConfigHandler,
	}
	cmd.Flags().StringP("provider", "p", "", "notification provider")
	cmd.Flags().StringToStringP("settings", "s", nil, "provider settings")
	return cmd
}
`

const exampleCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// New Command Group
func NewExampleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Example operations",
		Long:  "Manage example operations including creation, listing, and deletion",
	}

	cmd.AddCommand(
		newExampleCreateCommand(),
		newExampleListCommand(),
		newExampleDeleteCommand(),
	)

	return cmd
}

func newExampleCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new example",
		Args:  cobra.ExactArgs(1),
		RunE:  exampleCreateHandler,
	}
}

func newExampleListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List examples",
		RunE:  exampleListHandler,
	}
}

func newExampleDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete an example",
		Args:  cobra.ExactArgs(1),
		RunE:  exampleDeleteHandler,
	}
}
`

const migrationCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Migration Commands
func NewMigrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration operations",
		Long:  "Manage database migrations including running, rolling back, and creating migrations",
	}

	cmd.AddCommand(
		newMigrationRunCommand(),
		newMigrationRollbackCommand(),
		newMigrationCreateCommand(),
	)

	return cmd
}

func newMigrationRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run database migrations",
		RunE:  migrationRunHandler,
	}
}

func newMigrationRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last database migration",
		RunE:  migrationRollbackHandler,
	}
}

func newMigrationCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new database migration",
		Args:  cobra.ExactArgs(1),
		RunE:  migrationCreateHandler,
	}
}
`

const serverCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Server Commands
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server operations",
		Long:  "Manage server operations including starting and stopping the server",
	}

	cmd.AddCommand(
		newServerStartCommand(),
		newServerStopCommand(),
	)

	return cmd
}

func newServerStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		RunE:  serverStartHandler,
	}
}

func newServerStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the server",
		RunE:  serverStopHandler,
	}
}
`

const taskCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Task Commands
func NewTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task operations",
		Long:  "Manage task operations including running and listing tasks",
	}

	cmd.AddCommand(
		newTaskRunCommand(),
		newTaskListCommand(),
	)

	return cmd
}

func newTaskRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run [task_name]",
		Short: "Run a specific task",
		Args:  cobra.ExactArgs(1),
		RunE:  taskRunHandler,
	}
}

func newTaskListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available tasks",
		RunE:  taskListHandler,
	}
}
`

const searchCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Search Commands
func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search operations",
		Long:  "Manage search operations including indexing and querying",
	}

	cmd.AddCommand(
		newSearchIndexCommand(),
		newSearchQueryCommand(),
	)

	return cmd
}

func newSearchIndexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "index [document]",
		Short: "Index a new document",
		Args:  cobra.ExactArgs(1),
		RunE:  searchIndexHandler,
	}
}

func newSearchQueryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "query [term]",
		Short: "Query the search index",
		Args:  cobra.ExactArgs(1),
		RunE:  searchQueryHandler,
	}
}
`

const backupCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Backup Commands
func NewBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup operations",
		Long:  "Manage backup operations including creating, restoring, and listing backups",
	}

	cmd.AddCommand(
		newBackupCreateCommand(),
		newBackupRestoreCommand(),
		newBackupListCommand(),
	)

	return cmd
}

func newBackupCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new backup",
		Args:  cobra.ExactArgs(1),
		RunE:  backupCreateHandler,
	}
}

func newBackupRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [backup_id]",
		Short: "Restore from a backup",
		Args:  cobra.ExactArgs(1),
		RunE:  backupRestoreHandler,
	}
}

func newBackupListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List backups",
		RunE:  backupListHandler,
	}
}
`

const sessionCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Session Commands
func NewSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Session operations",
		Long:  "Manage user sessions including listing and terminating sessions",
	}

	cmd.AddCommand(
		newSessionListCommand(),
		newSessionTerminateCommand(),
	)

	return cmd
}

func newSessionListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE:  sessionListHandler,
	}
}

func newSessionTerminateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "terminate [session_id]",
		Short: "Terminate a session",
		Args:  cobra.ExactArgs(1),
		RunE:  sessionTerminateHandler,
	}
}
`

const maintenanceCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Maintenance Commands
func NewMaintenanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Maintenance operations",
		Long:  "Manage maintenance operations including starting and stopping maintenance mode",
	}

	cmd.AddCommand(
		newMaintenanceStartCommand(),
		newMaintenanceStopCommand(),
	)

	return cmd
}

func newMaintenanceStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start maintenance mode",
		RunE:  maintenanceStartHandler,
	}
}

func newMaintenanceStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop maintenance mode",
		RunE:  maintenanceStopHandler,
	}
}
`

const featureToggleCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Feature Toggle Commands
func NewFeatureToggleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature",
		Short: "Feature toggle operations",
		Long:  "Manage feature toggles including enabling and disabling features",
	}

	cmd.AddCommand(
		newFeatureEnableCommand(),
		newFeatureDisableCommand(),
	)

	return cmd
}

func newFeatureEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable [feature]",
		Short: "Enable a feature",
		Args:  cobra.ExactArgs(1),
		RunE:  featureEnableHandler,
	}
}

func newFeatureDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable [feature]",
		Short: "Disable a feature",
		Args:  cobra.ExactArgs(1),
		RunE:  featureDisableHandler,
	}
}
`

const schedulerCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Scheduler Commands
func NewSchedulerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Scheduler operations",
		Long:  "Manage scheduled tasks including creating, listing, and deleting schedules",
	}

	cmd.AddCommand(
		newSchedulerCreateCommand(),
		newSchedulerListCommand(),
		newSchedulerDeleteCommand(),
	)

	return cmd
}

func newSchedulerCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [task_name] [cron_expression]",
		Short: "Create a new scheduled task",
		Args:  cobra.ExactArgs(2),
		RunE:  schedulerCreateHandler,
	}
}

func newSchedulerListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled tasks",
		RunE:  schedulerListHandler,
	}
}

func newSchedulerDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [task_id]",
		Short: "Delete a scheduled task",
		Args:  cobra.ExactArgs(1),
		RunE:  schedulerDeleteHandler,
	}
}
`

const networkCommandsTemplate = `package commands

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/spf13/cobra"
)

// Network Commands
func NewNetworkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Network operations",
		Long:  "Manage network operations including configuration and monitoring",
	}

	cmd.AddCommand(
		newNetworkConfigCommand(),
		newNetworkMonitorCommand(),
		newNetworkEnvCheckCommand(),
		newNetworkPingCommand(),
		newNetworkTracerouteCommand(),
		newNetworkDNSLookupCommand(),
	)

	return cmd
}

func newNetworkConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure network settings",
		RunE:  networkConfigHandler,
	}
	cmd.Flags().StringP("ip", "i", "", "Set IP address")
	cmd.Flags().StringP("gateway", "g", "", "Set gateway")
	return cmd
}

func newNetworkMonitorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Monitor network traffic",
		RunE:  networkMonitorHandler,
	}
}

func newNetworkEnvCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "env-check",
		Short: "Check if running in Docker or bare-metal",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat("/.dockerenv"); err == nil {
				fmt.Println("Running inside Docker")
			} else if os.Getenv("DOCKER") != "" {
				fmt.Println("Running inside Docker")
			} else {
				fmt.Println("Running on bare-metal")
			}
			return nil
		},
	}
}

func newNetworkPingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ping [host]",
		Short: "Ping a network host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			var out []byte
			var err error

			if _, err := exec.LookPath("ping"); err == nil {
				out, err = exec.Command("ping", "-c", "4", host).Output()
			} else if _, err := exec.LookPath("fping"); err == nil {
				out, err = exec.Command("fping", "-c", "4", host).Output()
			} else {
				return fmt.Errorf("no suitable ping tool found")
			}

			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
}

func newNetworkTracerouteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "traceroute [host]",
		Short: "Perform a traceroute to a network host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			var out []byte
			var err error

			if _, err := exec.LookPath("traceroute"); err == nil {
				out, err = exec.Command("traceroute", host).Output()
			} else if _, err := exec.LookPath("tracepath"); err == nil {
				out, err = exec.Command("tracepath", host).Output()
			} else {
				return fmt.Errorf("no suitable traceroute tool found")
			}

			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
}

func newNetworkDNSLookupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dns-lookup [domain]",
		Short: "Perform a DNS lookup for a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			var out []byte
			var err error

			if _, err := exec.LookPath("nslookup"); err == nil {
				out, err = exec.Command("nslookup", domain).Output()
			} else if _, err := exec.LookPath("dig"); err == nil {
				out, err = exec.Command("dig", domain).Output()
			} else {
				return fmt.Errorf("no suitable DNS lookup tool found")
			}

			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
}
`

const licenseCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// License Commands
func NewLicenseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "License management operations",
		Long:  "Manage license operations including activation, deactivation, status checking, updating, and checking",
	}

	cmd.AddCommand(
		newLicenseActivateCommand(),
		newLicenseDeactivateCommand(),
		newLicenseStatusCommand(),
		newLicenseUpdateCommand(),
		newLicenseCheckCommand(),
	)

	return cmd
}

func newLicenseActivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "activate [license_key]",
		Short: "Activate a license",
		Args:  cobra.ExactArgs(1),
		RunE:  licenseActivateHandler,
	}
}

func newLicenseDeactivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deactivate [license_key]",
		Short: "Deactivate a license",
		Args:  cobra.ExactArgs(1),
		RunE:  licenseDeactivateHandler,
	}
}

func newLicenseStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status [license_key]",
		Short: "Check the status of a license",
		Args:  cobra.ExactArgs(1),
		RunE:  licenseStatusHandler,
	}
}

func newLicenseUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update [license_key]",
		Short: "Update a license",
		Args:  cobra.ExactArgs(1),
		RunE:  licenseUpdateHandler,
	}
}

func newLicenseCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check [license_key]",
		Short: "Check a license",
		Args:  cobra.ExactArgs(1),
		RunE:  licenseCheckHandler,
	}
}
`

const ldapCommandsTemplate = `package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// LDAP Commands
func NewLDAPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ldap",
		Short: "LDAP management operations",
		Long:  "Manage LDAP operations including connecting, searching, and modifying entries",
	}

	cmd.AddCommand(
		newLDAPConnectCommand(),
		newLDAPSearchCommand(),
		newLDAPModifyCommand(),
	)

	return cmd
}

func newLDAPConnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect [server]",
		Short: "Connect to an LDAP server",
		Args:  cobra.ExactArgs(1),
		RunE:  ldapConnectHandler,
	}
}

func newLDAPSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [base_dn] [filter]",
		Short: "Search LDAP entries",
		Args:  cobra.ExactArgs(2),
		RunE:  ldapSearchHandler,
	}
	cmd.Flags().StringSliceP("attributes", "a", []string{}, "attributes to retrieve")
	return cmd
}

func newLDAPModifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify [dn]",
		Short: "Modify an LDAP entry",
		Args:  cobra.ExactArgs(1),
		RunE:  ldapModifyHandler,
	}
	cmd.Flags().StringToStringP("attributes", "a", nil, "attributes to modify")
	return cmd
}
`
