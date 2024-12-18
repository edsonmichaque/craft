# Project Name

## Overview

This project uses the `urfave/cli` package to create a command-line interface (CLI) application. The CLI supports various commands and subcommands to perform different tasks. Additionally, it includes configuration parsing and follows best practices for error handling.

## Adding Subcommands

To add a subcommand to your CLI application, follow these steps:

1. **Define the Subcommand:**

   Create a new function that returns a `*cli.Command` struct. This struct should define the name, usage, flags, and action for the subcommand.

   ```go
   func CmdSubcommand(ctx context.Context, appCtx *Context) *cli.Command {
       return &cli.Command{
           Name:  "subcommand",
           Usage: "Description of the subcommand",
           Flags: []cli.Flag{
               // Define flags here
           },
           Action: func(c *cli.Context) error {
               // Implement the action for the subcommand
               return nil
           },
       }
   }
   ```

2. **Add the Subcommand to a Parent Command:**

   In the parent command's definition, add the subcommand to the `Subcommands` field.

   ```go
   func CmdParent(ctx context.Context, appCtx *Context) *cli.Command {
       return &cli.Command{
           Name:  "parent",
           Usage: "Description of the parent command",
           Subcommands: []*cli.Command{
               CmdSubcommand(ctx, appCtx),
               // Add more subcommands here
           },
       }
   }
   ```

3. **Register the Parent Command:**

   Ensure that the parent command is registered in your application's command list.

   ```go
   app := &cli.App{
       Commands: []*cli.Command{
           CmdParent(ctx, appCtx),
           // Other commands
       },
   }
   ```

## Adding Multilevel Subcommands

To add multilevel subcommands, follow the same pattern as above, but nest the subcommands within each other.

1. **Define the Nested Subcommand:**

   ```go
   func CmdNestedSubcommand(ctx context.Context, appCtx *Context) *cli.Command {
       return &cli.Command{
           Name:  "nested",
           Usage: "Description of the nested subcommand",
           Action: func(c *cli.Context) error {
               // Implement the action for the nested subcommand
               return nil
           },
       }
   }
   ```

2. **Add the Nested Subcommand to a Subcommand:**

   ```go
   func CmdSubcommand(ctx context.Context, appCtx *Context) *cli.Command {
       return &cli.Command{
           Name:  "subcommand",
           Usage: "Description of the subcommand",
           Subcommands: []*cli.Command{
               CmdNestedSubcommand(ctx, appCtx),
               // More nested subcommands
           },
       }
   }
   ```

3. **Register the Commands:**

   Ensure that all commands are registered in the application's command list as shown previously.

## Parsing Configurations

To parse configurations, you can use a configuration file or environment variables. Here's a basic example using a configuration file:

1. **Define a Configuration Struct:**

   ```go
   type Config struct {
       Port int    `json:"port"`
       Host string `json:"host"`
   }
   ```

2. **Load the Configuration:**

   Implement a function to load the configuration from a file.

   ```go
   func LoadConfig(filePath string) (*Config, error) {
       file, err := os.Open(filePath)
       if err != nil {
           return nil, err
       }
       defer file.Close()

       config := &Config{}
       decoder := json.NewDecoder(file)
       if err := decoder.Decode(config); err != nil {
           return nil, err
       }

       return config, nil
   }
   ```

3. **Use the Configuration in Commands:**

   Pass the configuration to your commands and use it as needed.

   ```go
   func CmdServer(ctx context.Context, appCtx *Context, config *Config) *cli.Command {
       // Use config.Port and config.Host
   }
   ```

## Best Practices

- **Avoid `os.Exit` in Commands/Subcommands:**

  It's a best practice to avoid calling `os.Exit` within commands or subcommands. Instead, return an error and handle it in the `main` function. This approach ensures that resources are properly cleaned up and allows for better error handling.

  ```go
  func main() {
      app := &cli.App{
          // Define commands
      }

      if err := app.Run(os.Args); err != nil {
          log.Fatal(err)
      }
  }
  ```

## Running the Application

To run the application, use the following command:

```bash
go run main.go [command] [subcommand] [flags]
```

Replace `[command]`, `[subcommand]`, and `[flags]` with the appropriate values for your use case.

## Conclusion

This template provides a basic structure for adding subcommands, parsing configurations, and following best practices in your CLI application using the `urfave/cli` package. Customize it further to fit your project's specific requirements.

## File Naming Conventions

To maintain a clear and organized project structure, it's recommended to use file names that reflect the command hierarchy. This approach helps in easily identifying the purpose of each file and its relation to the command structure.

1. **Top-Level Commands:**

   Files for top-level commands should be named after the command itself. 

   **Example:**
   - Command: `server`
   - File: `cmd_server.go`

2. **Subcommands:**

   For subcommands, include the parent command in the file name to indicate the hierarchy.

   **Example:**
   - Parent Command: `server`
   - Subcommand: `start`
   - File: `cmd_server_start.go`

3. **Nested Subcommands:**

   For nested subcommands, extend the file name to include all parent commands.

   **Example:**
   - Parent Command: `server`
   - Subcommand: `start`
   - Nested Subcommand: `restart`
   - File: `cmd_server_start_restart.go`

4. **Configuration Files:**

   Configuration files should be named to reflect their purpose.

   **Example:**
   - Configuration for server settings
   - File: `config_server.json`

By following these conventions, you ensure that the file structure mirrors the command hierarchy, making it easier to navigate and manage the codebase.