# Contributing to BBOX

Thank you for your interest in contributing to BBOX! We appreciate your efforts to improve the project. This guide will help you get started with your contributions.

## Table of Contents

1. [How to Report a Bug](#how-to-report-a-bug)
2. [How to Request a Feature](#how-to-request-a-feature)
3. [How to Suggest Documentation Improvements](#how-to-suggest-documentation-improvements)
4. [How to Submit a Pull Request](#how-to-submit-a-pull-request)
5. [Branch Naming Convention](#branch-naming-convention)
6. [Style Guide](#style-guide)
7. [Testing](#testing)
8. [Documentation](#documentation)
9. [Communication](#communication)

## How to Report a Bug

If you find a bug, please follow these steps:

1. **Search Existing Issues**: Before opening a new issue, please search the [existing issues](https://github.com/similarweb/bbox/issues) to avoid duplicates.
2. **Use the Bug Report Template**: If no similar issue exists, [open a new issue](https://github.com/similarweb/bbox/issues/new/choose) using the "Bug Report" template.
3. **Provide Detailed Information**: Include as much detail as possible, such as the environment, steps to reproduce the issue, and any relevant logs.

## How to Request a Feature

To suggest a new feature or enhancement:

1. **Check Existing Requests**: Browse the [existing issues](https://github.com/similarweb/bbox/issues) to see if a similar request has already been made.
2. **Use the Feature Request Template**: If your idea is unique, [open a new feature request](https://github.com/similarweb/bbox/issues/new/choose) using the "Feature Request" template.
3. **Describe Your Idea**: Provide a clear and concise description of your proposed feature, including its potential benefits.

## How to Suggest Documentation Improvements

To suggest changes or additions to the documentation:

1. **Check Existing Requests**: Browse the [existing issues](https://github.com/similarweb/bbox/issues) to see if a similar request has already been made.
2. **Use the Documentation Request Template**: If you have ideas on how to improve the documentation, [open a new documentation request](https://github.com/similarweb/bbox/issues/new/choose) using the "Documentation Request" template.
3. **Provide Specific Suggestions**: Clearly specify which part of the documentation needs improvement and offer your suggestions for what could be added or changed.
4. **Add Context**: Include any additional context or examples that might help clarify your request.

Your feedback helps us make the documentation better for everyone!

## How to Submit a Pull Request

To contribute code:

1. **Fork the Repository**: Create a personal fork of the repository on GitHub.
2. **Clone Your Fork**: Clone your fork to your local machine.
3. **Create a New Branch**: Follow the branch naming convention provided in the [Branch Naming Convention](#branch-naming-convention) section.
4. **Make Your Changes**: Implement your changes in your branch.
5. **Run Linter**: Ensure your code passes `golangci-lint`.
6. **Test Your Changes**: Write tests for your code and ensure all tests pass.
7. **Submit a Pull Request**: Push your branch to GitHub and [open a pull request](https://github.com/similarweb/bbox/compare) against the main branch.

## Branch Naming Convention

We follow a specific convention for commit messages to maintain consistency and clarity. Please use the following format:

```bash
[type]/[scope]_Summary
```

- **Type**: The type of change (e.g., feat, fix, improve, cleanup, refactor, revert).
- **Scope**: The scope of the change (e.g., admin, cli, docker, multi-trigger, test, ci, build, version, doc, auth).
- **Summary**: A brief, self-explanatory description of the change in present tense imperative, starting with a capital letter and no period at the end.

For example:

```bash
feat/clean_add-new-clean-feature
```

## Style Guide

Please adhere to the following coding standards:

- **Language**: The project is written in Go. Follow Go conventions and best practices.
- **Linting**: Please ensure that your code adheres to the project's linting standards. Run `golangci-lint` before submitting your pull request to ensure code quality and consistency across the codebase.

## Testing

We have a tests pipeline in place to automatically run tests against your code changes. You don't need to run the tests manually, but you must ensure that your code adheres to the project's testing standards and does not introduce any issues.

### How to Write Proper Tests

To maintain the quality and reliability of our codebase, we require that all new code contributions include corresponding tests. Here are some guidelines to help you write effective tests:

1. Use Descriptive Test Names:

    - Each test should have a clear and descriptive name that indicates what the test is verifying.
    - Example: TestTrigger_SuccessfulTriggerWithoutWait is better than TestTrigger.

2. Structure Your Tests:

    - Use t.Run() to create subtests for different scenarios within the same test function.

    - Organize your test data in a struct to handle various input and expected output cases.

3. Mock External Dependencies:

    - When writing tests for functions that interact with external services, use mocking to simulate those services.

    - Utilize the testutils package to create mock services, as shown in the example below.

4. Test a Range of Scenarios:

    - Your tests should cover different scenarios, including success, failure, and edge cases.

    - Ensure that you test with both typical and extreme inputs.

5. Assert Expectations:

    - Use AssertExpectations() to ensure that all mocked methods were called with the expected parameters and that the expected outcomes were met.

6. Example Test Structure:

    ```go
    func TestTrigger(t *testing.T) {
        tests := []struct {
            name        string
            buildTypeID string
            branchName  string
            // Other fields...
        }{
            {
                name:        "Successful Trigger without Wait",
                buildTypeID: "bt123",
                branchName:  "master",
                // Set other fields...
            },
            // Add more test cases...
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                mockBuild := new(testutils.MockBuildService)
                mockArtifacts := new(testutils.MockArtifactsService)

                client := &teamcity.Client{
                    Build:     mockBuild,
                    Artifacts: mockArtifacts,
                }

                // Set expectations on the mocks
                mockBuild.On("TriggerBuild", tt.buildTypeID, tt.branchName, tt.properties).Return(tt.triggerBuildResponse, tt.waitForBuildError)
                
                // Call the function under test
                trigger(client, tt.buildTypeID, tt.branchName, tt.artifactsPath, tt.properties, tt.requireArtifacts, tt.waitForBuild, tt.downloadArtifacts, tt.waitForBuildTimeout)

                // Assert that the expectations were met
                mockBuild.AssertExpectations(t)
                mockArtifacts.AssertExpectations(t)
            })
        }
    }
    ```

7. Running Tests:

    - You don't need to manually run the tests; our CI pipeline will automatically execute them whenever you push changes or create a pull request.

    - However, you can run the tests locally using go test ./... -v if you want to verify them before pushing.

By following these guidelines, you'll help ensure that your contributions are robust, reliable, and integrate smoothly with the existing codebase.

## Documentation

Please update the documentation to reflect any changes you make to the codebase:

- **Code Documentation**: Include inline comments and function/method docstrings as needed.
- **Project Documentation**: Update the `README.md` or other relevant docs with information about new features or updates.

### Documentation for Cobra Commands

When adding or updating Cobra commands in BBOX, it is important to include both short and long descriptions to provide clarity for users:

- **Short Description**: This should be a concise, one-sentence summary of the command's purpose. It is displayed in the list of available commands and when users request help for a specific command.

- **Long Description**: This should provide a more detailed explanation of what the command does, including any important context, usage notes, or examples. The long description is displayed when users request detailed help for a command.

#### Example Structure

```go
var exampleCmd = &cobra.Command{
 Use:   "example",
 Short: "Briefly describe what the example command does",
 Long:  `This is a more detailed description of the example command, explaining its purpose, how it works, and any other relevant information. This can include usage notes, examples, and any warnings or important considerations.`,
 Run: func(cmd *cobra.Command, args []string) {
  // Command implementation
 },
}
```

## Communication

We encourage contributors to discuss significant changes or ideas before submitting them. Use [GitHub Discussions](https://github.com/similarweb/bbox/discussions) for any general questions, ideas, or to seek feedback before creating an issue or pull request.

---

Thank you for contributing! Your support and involvement are what make BBOX better.
