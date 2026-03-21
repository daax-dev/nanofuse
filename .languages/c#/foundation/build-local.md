# C# Local Build System & Development Foundation

## Table of Contents
1. [.NET SDK Installation and Management](#net-sdk-installation-and-management)
2. [Project Templates and Creation](#project-templates-and-creation)
3. [Solution and Project Structure](#solution-and-project-structure)
4. [Build, Restore, and Publish Commands](#build-restore-and-publish-commands)
5. [Configuration Files](#configuration-files)
6. [MSBuild Fundamentals](#msbuild-fundamentals)
7. [Package Management with NuGet](#package-management-with-nuget)
8. [Testing with dotnet test](#testing-with-dotnet-test)
9. [Code Quality Tools](#code-quality-tools)
10. [Performance Profiling and Diagnostics](#performance-profiling-and-diagnostics)
11. [Docker Integration](#docker-integration)
12. [Advanced Build Scenarios](#advanced-build-scenarios)

---

## .NET SDK Installation and Management

### Multi-Version SDK Management

#### Installing Multiple SDKs
```bash
# Download and install from official site
# https://dotnet.microsoft.com/download

# Check installed SDKs
dotnet --list-sdks
# Output:
# 6.0.417 [/usr/share/dotnet/sdk]
# 7.0.404 [/usr/share/dotnet/sdk]
# 8.0.100 [/usr/share/dotnet/sdk]

# Check runtimes
dotnet --list-runtimes
```

#### Global.json for Version Control
```json
{
  "sdk": {
    "version": "8.0.100",
    "rollForward": "latestMinor",
    "allowPrerelease": false
  },
  "msbuild-sdks": {
    "Microsoft.Build.NoTargets": "3.7.0"
  }
}
```

#### SDK Version Selection Logic
```bash
# Project-level global.json takes precedence
# Falls back to latest compatible version
# Roll-forward policies:
# - patch: 8.0.100 → 8.0.101
# - feature: 8.0.100 → 8.0.200
# - minor: 8.0.100 → 8.1.100
# - major: 8.0.100 → 9.0.100
# - latestPatch, latestFeature, latestMinor, latestMajor
```

### Environment Setup

#### Development Environment Configuration
```bash
# Set environment variables
export DOTNET_CLI_TELEMETRY_OPTOUT=1
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1
export DOTNET_NOLOGO=1
export DOTNET_GENERATE_ASPNET_CERTIFICATE=false

# NuGet configuration
export NUGET_PACKAGES=/path/to/custom/packages
export DOTNET_NUGET_SIGNATURE_VERIFICATION=false
```

#### IDE Integration
```bash
# VS Code extensions
code --install-extension ms-dotnettools.csharp
code --install-extension ms-dotnettools.vscode-dotnet-runtime

# JetBrains Rider
# Built-in .NET support with advanced debugging
```

---

## Project Templates and Creation

### Built-in Templates

#### Common Project Templates
```bash
# List all templates
dotnet new list

# Console application
dotnet new console -n MyConsoleApp
dotnet new console -n MyConsoleApp -f net8.0 --use-program-main

# Class library
dotnet new classlib -n MyLibrary

# Web API
dotnet new webapi -n MyWebApi
dotnet new webapi -n MyWebApi --use-controllers

# MVC Web App
dotnet new mvc -n MyMvcApp

# Blazor applications
dotnet new blazor -n MyBlazorApp
dotnet new blazorserver -n MyBlazorServerApp

# Test projects
dotnet new xunit -n MyLibrary.Tests
dotnet new nunit -n MyLibrary.NUnit.Tests
dotnet new mstest -n MyLibrary.MSTest.Tests

# Worker service
dotnet new worker -n MyWorkerService
```

#### Template Parameters
```bash
# Framework targeting
dotnet new console -f net8.0
dotnet new console -f net48
dotnet new console -f net6.0

# Language selection
dotnet new console -lang C#
dotnet new console -lang F#
dotnet new console -lang VB

# Other common parameters
dotnet new webapi --no-https --no-openapi
dotnet new console --use-program-main
dotnet new classlib --framework net8.0 --langVersion latest
```

### Custom Templates

#### Creating Custom Templates
```bash
# Template structure
MyTemplate/
├── .template.config/
│   └── template.json
├── MyTemplate.csproj
├── Program.cs
└── README.md
```

#### Template Configuration (template.json)
```json
{
  "$schema": "http://json.schemastore.org/template",
  "author": "Your Name",
  "classifications": ["Console", "Custom"],
  "identity": "Custom.Console.Template",
  "name": "Custom Console App",
  "shortName": "customconsole",
  "sourceName": "MyTemplate",
  "preferNameDirectory": true,
  "tags": {
    "language": "C#",
    "type": "project"
  },
  "symbols": {
    "Framework": {
      "type": "parameter",
      "description": "The target framework",
      "datatype": "choice",
      "choices": [
        {
          "choice": "net8.0",
          "description": ".NET 8"
        },
        {
          "choice": "net6.0",
          "description": ".NET 6"
        }
      ],
      "defaultValue": "net8.0",
      "replaces": "FRAMEWORK_VERSION"
    }
  }
}
```

#### Installing and Using Custom Templates
```bash
# Install from local folder
dotnet new install ./MyTemplate

# Install from NuGet package
dotnet new install MyCompany.Templates

# Uninstall template
dotnet new uninstall MyCompany.Templates

# Use custom template
dotnet new customconsole -n MyNewProject
```

---

## Solution and Project Structure

### Solution Management

#### Creating and Managing Solutions
```bash
# Create new solution
dotnet new sln -n MySolution

# Add projects to solution
dotnet sln add src/MyLibrary/MyLibrary.csproj
dotnet sln add tests/MyLibrary.Tests/MyLibrary.Tests.csproj
dotnet sln add src/MyWebApi/MyWebApi.csproj

# Remove project from solution
dotnet sln remove src/MyLibrary/MyLibrary.csproj

# List projects in solution
dotnet sln list
```

#### Recommended Solution Structure
```
MySolution/
├── MySolution.sln
├── global.json
├── Directory.Build.props
├── Directory.Build.targets
├── Directory.Packages.props
├── .editorconfig
├── .gitignore
├── README.md
├── src/
│   ├── MyLibrary/
│   │   ├── MyLibrary.csproj
│   │   └── *.cs files
│   └── MyWebApi/
│       ├── MyWebApi.csproj
│       └── *.cs files
├── tests/
│   ├── MyLibrary.Tests/
│   │   ├── MyLibrary.Tests.csproj
│   │   └── *.cs test files
│   └── MyLibrary.IntegrationTests/
├── docs/
├── samples/
├── tools/
└── build/
    ├── build.props
    ├── build.targets
    └── scripts/
```

### Project References

#### Adding Project References
```bash
# Add project reference
dotnet add src/MyWebApi reference src/MyLibrary

# Add package reference
dotnet add src/MyWebApi package Microsoft.EntityFrameworkCore

# Add package with version
dotnet add package Newtonsoft.Json --version 13.0.3

# Remove reference
dotnet remove src/MyWebApi reference src/MyLibrary
dotnet remove src/MyWebApi package Newtonsoft.Json
```

#### Project File Structure (.csproj)
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <LangVersion>latest</LangVersion>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <WarningsAsErrors />
    <WarningsNotAsErrors>CS1591</WarningsNotAsErrors>
  </PropertyGroup>

  <PropertyGroup Condition="'$(Configuration)'=='Release'">
    <Optimize>true</Optimize>
    <DebugType>portable</DebugType>
    <DebugSymbols>true</DebugSymbols>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="8.0.0" />
    <PackageReference Include="Serilog.Extensions.Hosting" Version="8.0.0" />
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="../MyLibrary/MyLibrary.csproj" />
  </ItemGroup>

  <ItemGroup>
    <None Include="appsettings.json" CopyToOutputDirectory="PreserveNewest" />
    <Content Include="wwwroot/**" CopyToOutputDirectory="PreserveNewest" />
  </ItemGroup>

</Project>
```

---

## Build, Restore, and Publish Commands

### Build Commands

#### Basic Build Operations
```bash
# Restore packages (implicit in build)
dotnet restore

# Build project/solution
dotnet build
dotnet build --configuration Release
dotnet build --no-restore
dotnet build --verbosity detailed

# Clean build artifacts
dotnet clean
dotnet clean --configuration Release

# Rebuild (clean + build)
dotnet build --no-incremental
```

#### Advanced Build Options
```bash
# Multi-targeting build
dotnet build --framework net8.0
dotnet build --framework net48

# Runtime-specific build
dotnet build --runtime win-x64
dotnet build --runtime linux-x64
dotnet build --runtime osx-x64

# Build with MSBuild properties
dotnet build -p:Configuration=Release -p:Platform=x64
dotnet build -p:DefineConstants="DEBUG;TRACE;CUSTOM"

# Parallel builds
dotnet build --maxcpucount:4
dotnet build -m:4
```

### Publish Commands

#### Application Publishing
```bash
# Framework-dependent deployment
dotnet publish --configuration Release --output ./publish

# Self-contained deployment
dotnet publish --configuration Release --runtime win-x64 --self-contained
dotnet publish -c Release -r linux-x64 --self-contained true

# Single file deployment
dotnet publish -c Release -r win-x64 --self-contained \
  -p:PublishSingleFile=true -p:PublishTrimmed=true

# ReadyToRun images
dotnet publish -c Release -r win-x64 --self-contained \
  -p:PublishReadyToRun=true
```

#### Publishing Profiles
```xml
<!-- Properties/PublishProfiles/Production.pubxml -->
<Project>
  <PropertyGroup>
    <PublishProtocol>FileSystem</PublishProtocol>
    <Configuration>Release</Configuration>
    <Platform>Any CPU</Platform>
    <TargetFramework>net8.0</TargetFramework>
    <PublishDir>bin\Release\net8.0\publish\</PublishDir>
    <PublishUrl>bin\Release\net8.0\publish\</PublishUrl>
    <SelfContained>false</SelfContained>
    <PublishSingleFile>false</PublishSingleFile>
    <PublishReadyToRun>false</PublishReadyToRun>
  </PropertyGroup>
</Project>
```

```bash
# Use publishing profile
dotnet publish -p:PublishProfile=Production
```

### Run Commands

#### Development Execution
```bash
# Run project
dotnet run
dotnet run --project src/MyWebApi

# Run with arguments
dotnet run -- arg1 arg2 --option value

# Run specific configuration
dotnet run --configuration Release

# Watch for changes and restart
dotnet watch run
dotnet watch --project src/MyWebApi run
```

---

## Configuration Files

### Directory.Build.props

#### Central Property Management
```xml
<Project>

  <!-- Version and metadata -->
  <PropertyGroup>
    <Company>My Company</Company>
    <Product>My Product Suite</Product>
    <Copyright>Copyright © My Company 2024</Copyright>
    <Version>1.0.0</Version>
    <AssemblyVersion>1.0.0.0</AssemblyVersion>
    <FileVersion>1.0.0.0</FileVersion>
  </PropertyGroup>

  <!-- Global compiler settings -->
  <PropertyGroup>
    <LangVersion>latest</LangVersion>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <WarningsAsErrors />
    <WarningsNotAsErrors>CS1591;CS0618</WarningsNotAsErrors>
    <NoWarn>$(NoWarn);CS1591</NoWarn>
  </PropertyGroup>

  <!-- Debug/Release configurations -->
  <PropertyGroup Condition="'$(Configuration)'=='Debug'">
    <DebugType>full</DebugType>
    <DebugSymbols>true</DebugSymbols>
    <Optimize>false</Optimize>
    <DefineConstants>DEBUG;TRACE</DefineConstants>
  </PropertyGroup>

  <PropertyGroup Condition="'$(Configuration)'=='Release'">
    <DebugType>portable</DebugType>
    <DebugSymbols>true</DebugSymbols>
    <Optimize>true</Optimize>
    <DefineConstants>RELEASE;TRACE</DefineConstants>
  </PropertyGroup>

  <!-- Code analysis -->
  <PropertyGroup>
    <EnableNETAnalyzers>true</EnableNETAnalyzers>
    <AnalysisLevel>latest</AnalysisLevel>
    <AnalysisMode>Recommended</AnalysisMode>
    <CodeAnalysisRuleSet>$(MSBuildThisFileDirectory)CodeAnalysis.ruleset</CodeAnalysisRuleSet>
  </PropertyGroup>

  <!-- Global package references -->
  <ItemGroup>
    <PackageReference Include="Microsoft.SourceLink.GitHub" Version="8.0.0" PrivateAssets="All" />
  </ItemGroup>

  <!-- Conditional compilation symbols by target framework -->
  <PropertyGroup Condition="'$(TargetFramework)'=='net8.0'">
    <DefineConstants>$(DefineConstants);NET8_0_OR_GREATER</DefineConstants>
  </PropertyGroup>

</Project>
```

### Directory.Build.targets

#### Post-build Customizations
```xml
<Project>

  <!-- Custom build tasks -->
  <Target Name="PrintBuildInfo" BeforeTargets="Build">
    <Message Text="Building $(MSBuildProjectName) for $(TargetFramework)" Importance="high" />
    <Message Text="Output: $(OutputPath)" Importance="normal" />
  </Target>

  <!-- Copy additional files -->
  <Target Name="CopyConfigFiles" AfterTargets="Build">
    <ItemGroup>
      <ConfigFiles Include="$(MSBuildProjectDirectory)\configs\*.json" />
    </ItemGroup>
    <Copy SourceFiles="@(ConfigFiles)" DestinationFolder="$(OutDir)configs" />
  </Target>

  <!-- Version stamping -->
  <Target Name="SetVersionInfo" BeforeTargets="GetAssemblyVersion">
    <PropertyGroup>
      <BuildNumber Condition="'$(BUILD_NUMBER)'!=''">$(BUILD_NUMBER)</BuildNumber>
      <BuildNumber Condition="'$(BUILD_NUMBER)'==''">0</BuildNumber>
      <AssemblyVersion>$(MajorVersion).$(MinorVersion).$(PatchVersion).$(BuildNumber)</AssemblyVersion>
    </PropertyGroup>
  </Target>

</Project>
```

### Directory.Packages.props

#### Central Package Management (CPM)
```xml
<Project>

  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
    <CentralPackageTransitivePinningEnabled>true</CentralPackageTransitivePinningEnabled>
  </PropertyGroup>

  <ItemGroup>
    <!-- Microsoft packages -->
    <PackageVersion Include="Microsoft.Extensions.Hosting" Version="8.0.0" />
    <PackageVersion Include="Microsoft.Extensions.DependencyInjection" Version="8.0.0" />
    <PackageVersion Include="Microsoft.Extensions.Configuration" Version="8.0.0" />
    <PackageVersion Include="Microsoft.Extensions.Logging" Version="8.0.0" />

    <!-- Entity Framework -->
    <PackageVersion Include="Microsoft.EntityFrameworkCore" Version="8.0.0" />
    <PackageVersion Include="Microsoft.EntityFrameworkCore.SqlServer" Version="8.0.0" />
    <PackageVersion Include="Microsoft.EntityFrameworkCore.Tools" Version="8.0.0" />

    <!-- ASP.NET Core -->
    <PackageVersion Include="Microsoft.AspNetCore.Authentication.JwtBearer" Version="8.0.0" />
    <PackageVersion Include="Microsoft.AspNetCore.OpenApi" Version="8.0.0" />
    <PackageVersion Include="Swashbuckle.AspNetCore" Version="6.5.0" />

    <!-- Third-party packages -->
    <PackageVersion Include="Newtonsoft.Json" Version="13.0.3" />
    <PackageVersion Include="Serilog.Extensions.Hosting" Version="8.0.0" />
    <PackageVersion Include="AutoMapper" Version="12.0.1" />

    <!-- Testing packages -->
    <PackageVersion Include="Microsoft.NET.Test.Sdk" Version="17.8.0" />
    <PackageVersion Include="xunit" Version="2.6.2" />
    <PackageVersion Include="xunit.runner.visualstudio" Version="2.5.3" />
    <PackageVersion Include="FluentAssertions" Version="6.12.0" />
    <PackageVersion Include="Moq" Version="4.20.69" />
    <PackageVersion Include="Microsoft.AspNetCore.Mvc.Testing" Version="8.0.0" />
  </ItemGroup>

</Project>
```

### EditorConfig

#### Code Style Configuration
```ini
# .editorconfig

root = true

[*]
charset = utf-8
end_of_line = crlf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2

[*.{cs,csx,cake}]
indent_size = 4

# C# formatting rules
[*.cs]
# Organize usings
dotnet_sort_system_directives_first = true
dotnet_separate_import_directive_groups = false

# this. preferences
dotnet_style_qualification_for_field = false:suggestion
dotnet_style_qualification_for_property = false:suggestion
dotnet_style_qualification_for_method = false:suggestion
dotnet_style_qualification_for_event = false:suggestion

# Language keywords vs BCL types preferences
dotnet_style_predefined_type_for_locals_parameters_members = true:suggestion
dotnet_style_predefined_type_for_member_access = true:suggestion

# Parentheses preferences
dotnet_style_parentheses_in_arithmetic_binary_operators = always_for_clarity:silent
dotnet_style_parentheses_in_relational_binary_operators = always_for_clarity:silent

# Modifier preferences
dotnet_style_require_accessibility_modifiers = always:suggestion
dotnet_style_readonly_field = true:suggestion

# Expression-level preferences
dotnet_style_object_initializer = true:suggestion
dotnet_style_collection_initializer = true:suggestion
dotnet_style_explicit_tuple_names = true:suggestion
dotnet_style_null_propagation = true:suggestion

# C# Code style rules
csharp_new_line_before_open_brace = all
csharp_new_line_before_else = true
csharp_new_line_before_catch = true
csharp_new_line_before_finally = true
csharp_new_line_before_members_in_object_initializers = true
csharp_new_line_before_members_in_anonymous_types = true

# Indentation preferences
csharp_indent_case_contents = true
csharp_indent_switch_labels = true

# Space preferences
csharp_space_after_cast = false
csharp_space_after_keywords_in_control_flow_statements = true
csharp_space_around_binary_operators = before_and_after

# var preferences
csharp_style_var_for_built_in_types = false:suggestion
csharp_style_var_when_type_is_apparent = true:suggestion
csharp_style_var_elsewhere = false:suggestion

[*.{json,yml,yaml}]
indent_size = 2

[*.{props,targets,proj,projitems,shproj}]
indent_size = 2
```

---

## MSBuild Fundamentals

### MSBuild Project Structure

#### Basic MSBuild Concepts
```xml
<!-- MSBuild evaluation order:
     1. Imports (Sdk, Directory.Build.props)
     2. Properties
     3. Items
     4. Targets
     5. Directory.Build.targets
-->

<Project Sdk="Microsoft.NET.Sdk">

  <!-- Properties: Key-value pairs -->
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <OutputType>Exe</OutputType>
    <CustomProperty>CustomValue</CustomProperty>
  </PropertyGroup>

  <!-- Items: Lists of files or other things -->
  <ItemGroup>
    <PackageReference Include="Microsoft.Extensions.Hosting" Version="8.0.0" />
    <Compile Include="**/*.cs" Exclude="bin/**;obj/**" />
    <Content Include="appsettings.json" CopyToOutputDirectory="PreserveNewest" />
  </ItemGroup>

  <!-- Targets: Build steps -->
  <Target Name="CustomTarget" BeforeTargets="Build">
    <Message Text="Executing custom target" Importance="high" />
  </Target>

</Project>
```

#### Property Functions and Conditions
```xml
<Project>

  <PropertyGroup>
    <!-- Property functions -->
    <CurrentYear>$([System.DateTime]::Now.Year)</CurrentYear>
    <UpperCaseProjectName>$(MSBuildProjectName.ToUpper())</UpperCaseProjectName>
    <IsWindows Condition="'$([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform($([System.Runtime.InteropServices.OSPlatform]::Windows)))' == 'true'">true</IsWindows>

    <!-- Conditional properties -->
    <OutputPath Condition="'$(Configuration)'=='Debug'">bin\Debug\</OutputPath>
    <OutputPath Condition="'$(Configuration)'=='Release'">bin\Release\</OutputPath>

    <!-- String manipulation -->
    <ModifiedString>$([System.String]::Copy('$(SomeProperty)').Replace('old', 'new'))</ModifiedString>
  </PropertyGroup>

  <!-- Conditional items -->
  <ItemGroup Condition="'$(TargetFramework)'=='net8.0'">
    <PackageReference Include="Net8SpecificPackage" Version="1.0.0" />
  </ItemGroup>

</Project>
```

### Custom MSBuild Tasks

#### Creating Custom Tasks
```csharp
// CustomTask.cs
using Microsoft.Build.Framework;
using Microsoft.Build.Utilities;

public class GenerateVersionInfo : Task
{
    [Required]
    public string OutputFile { get; set; }

    [Required]
    public string Version { get; set; }

    public override bool Execute()
    {
        try
        {
            var content = $@"
using System.Reflection;

[assembly: AssemblyVersion(""{Version}"")]
[assembly: AssemblyFileVersion(""{Version}"")]
[assembly: AssemblyInformationalVersion(""{Version}"")]
";
            File.WriteAllText(OutputFile, content);
            Log.LogMessage(MessageImportance.High, $"Generated version info: {OutputFile}");
            return true;
        }
        catch (Exception ex)
        {
            Log.LogError($"Error generating version info: {ex.Message}");
            return false;
        }
    }
}
```

#### Using Custom Tasks
```xml
<!-- Build.targets -->
<Project>

  <UsingTask TaskName="GenerateVersionInfo" AssemblyFile="$(MSBuildThisFileDirectory)tasks\CustomTasks.dll" />

  <Target Name="GenerateVersionInfo" BeforeTargets="PrepareForBuild">
    <GenerateVersionInfo
      OutputFile="$(MSBuildProjectDirectory)\Properties\VersionInfo.cs"
      Version="$(Version)" />
    <ItemGroup>
      <Compile Include="Properties\VersionInfo.cs" />
    </ItemGroup>
  </Target>

</Project>
```

### MSBuild Debugging

#### Diagnostic Tools
```bash
# Binary log for detailed analysis
dotnet build -bl:build.binlog

# View binary log
# Install: dotnet tool install --global MSBuildStructuredLog
MSBuildStructuredLogViewer build.binlog

# Verbose logging
dotnet build --verbosity diagnostic > build.log

# Property and item evaluation
dotnet build -pp:evaluated.proj  # Preprocessed project
dotnet build -targets           # List available targets
```

---

## Package Management with NuGet

### Package Management Commands

#### Basic Package Operations
```bash
# Search packages
dotnet nuget search EntityFramework
nuget search EntityFramework

# Add package references
dotnet add package Microsoft.EntityFrameworkCore
dotnet add package Microsoft.EntityFrameworkCore --version 8.0.0
dotnet add package Microsoft.EntityFrameworkCore --source https://api.nuget.org/v3/index.json

# Update packages
dotnet add package Microsoft.EntityFrameworkCore  # Updates to latest
dotnet list package --outdated
dotnet list package --include-transitive

# Remove packages
dotnet remove package Microsoft.EntityFrameworkCore
```

#### Package Sources and Configuration
```bash
# List package sources
dotnet nuget list source

# Add custom source
dotnet nuget add source https://pkgs.dev.azure.com/myorg/_packaging/myfeed/nuget/v3/index.json --name MyOrgFeed

# Remove source
dotnet nuget remove source MyOrgFeed

# Update source
dotnet nuget update source MyOrgFeed --source https://new-url/v3/index.json
```

### NuGet.Config

#### Configuration Hierarchy
```xml
<!-- %AppData%\NuGet\NuGet.Config (User level) -->
<!-- Solution root NuGet.Config (Solution level) -->
<!-- Project folder NuGet.Config (Project level) -->

<?xml version="1.0" encoding="utf-8"?>
<configuration>

  <config>
    <add key="globalPackagesFolder" value="C:\packages" />
    <add key="repositoryPath" value=".\packages" />
    <add key="defaultPushSource" value="https://api.nuget.org/v3/index.json" />
    <add key="http_proxy" value="http://proxy.company.com:8080" />
    <add key="signatureValidationMode" value="require" />
  </config>

  <packageSources>
    <clear />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
    <add key="MyOrgFeed" value="https://pkgs.dev.azure.com/myorg/_packaging/myfeed/nuget/v3/index.json" />
    <add key="LocalFeed" value="C:\LocalPackages" />
  </packageSources>

  <packageSourceMapping>
    <packageSource key="nuget.org">
      <package pattern="*" />
    </packageSource>
    <packageSource key="MyOrgFeed">
      <package pattern="MyCompany.*" />
    </packageSource>
  </packageSourceMapping>

  <packageSourceCredentials>
    <MyOrgFeed>
      <add key="Username" value="%NUGET_USERNAME%" />
      <add key="ClearTextPassword" value="%NUGET_PASSWORD%" />
    </MyOrgFeed>
  </packageSourceCredentials>

  <trustedSigners>
    <author name="Microsoft">
      <certificate fingerprint="3F9001EA83C560D712C24CF213C3D312CB3BFF51EE89435D3430BD06B5D0EECE"
                   hashAlgorithm="SHA256"
                   allowUntrustedRoot="false" />
    </author>
  </trustedSigners>

</configuration>
```

### Creating NuGet Packages

#### Package Project Configuration
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <GeneratePackageOnBuild>true</GeneratePackageOnBuild>

    <!-- Package metadata -->
    <PackageId>MyCompany.MyLibrary</PackageId>
    <PackageVersion>1.2.3</PackageVersion>
    <Authors>My Name</Authors>
    <Company>My Company</Company>
    <Description>A useful library for doing things</Description>
    <PackageLicenseExpression>MIT</PackageLicenseExpression>
    <PackageProjectUrl>https://github.com/myorg/mylibrary</PackageProjectUrl>
    <PackageIcon>icon.png</PackageIcon>
    <PackageReadmeFile>README.md</PackageReadmeFile>
    <RepositoryUrl>https://github.com/myorg/mylibrary.git</RepositoryUrl>
    <RepositoryType>git</RepositoryType>
    <PackageTags>library;utility;awesome</PackageTags>
    <PackageReleaseNotes>Bug fixes and performance improvements</PackageReleaseNotes>

    <!-- Advanced packaging -->
    <IncludeSymbols>true</IncludeSymbols>
    <SymbolPackageFormat>snupkg</SymbolPackageFormat>
    <PublishRepositoryUrl>true</PublishRepositoryUrl>
    <EmbedUntrackedSources>true</EmbedUntrackedSources>
  </PropertyGroup>

  <ItemGroup>
    <None Include="README.md" Pack="true" PackagePath="" />
    <None Include="icon.png" Pack="true" PackagePath="" />
    <None Include="LICENSE" Pack="true" PackagePath="" />
  </ItemGroup>

  <!-- Include additional files -->
  <ItemGroup>
    <Content Include="content\**" Pack="true" PackagePath="content\" />
    <None Include="tools\**" Pack="true" PackagePath="tools\" />
  </ItemGroup>

</Project>
```

#### Creating and Publishing Packages
```bash
# Create package
dotnet pack
dotnet pack --configuration Release --output nupkgs

# Publish to NuGet.org
dotnet nuget push MyLibrary.1.0.0.nupkg --api-key YOUR_API_KEY --source https://api.nuget.org/v3/index.json

# Publish to private feed
dotnet nuget push MyLibrary.1.0.0.nupkg --source MyOrgFeed

# Delete package (within 72 hours)
dotnet nuget delete MyLibrary 1.0.0 --api-key YOUR_API_KEY --source https://api.nuget.org/v3/index.json
```

---

## Testing with dotnet test

### Test Project Setup

#### Test Project Configuration
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <IsPackable>false</IsPackable>
    <IsTestProject>true</IsTestProject>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.NET.Test.Sdk" />
    <PackageReference Include="xunit" />
    <PackageReference Include="xunit.runner.visualstudio" />
    <PackageReference Include="coverlet.collector" />
    <PackageReference Include="FluentAssertions" />
    <PackageReference Include="Moq" />
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="..\MyLibrary\MyLibrary.csproj" />
  </ItemGroup>

</Project>
```

#### Sample Test Class
```csharp
using FluentAssertions;
using Moq;
using Xunit;
using Xunit.Abstractions;

namespace MyLibrary.Tests
{
    public class CalculatorTests
    {
        private readonly ITestOutputHelper _output;

        public CalculatorTests(ITestOutputHelper output)
        {
            _output = output;
        }

        [Fact]
        public void Add_TwoPositiveNumbers_ReturnsSum()
        {
            // Arrange
            var calculator = new Calculator();

            // Act
            var result = calculator.Add(2, 3);

            // Assert
            result.Should().Be(5);
            _output.WriteLine($"Result: {result}");
        }

        [Theory]
        [InlineData(1, 2, 3)]
        [InlineData(0, 0, 0)]
        [InlineData(-1, 1, 0)]
        public void Add_VariousInputs_ReturnsCorrectSum(int a, int b, int expected)
        {
            // Arrange
            var calculator = new Calculator();

            // Act
            var result = calculator.Add(a, b);

            // Assert
            result.Should().Be(expected);
        }

        [Fact]
        public async Task GetDataAsync_ValidInput_ReturnsData()
        {
            // Arrange
            var mockService = new Mock<IDataService>();
            mockService.Setup(s => s.GetAsync(It.IsAny<int>()))
                      .ReturnsAsync("test data");

            var sut = new DataProcessor(mockService.Object);

            // Act
            var result = await sut.ProcessAsync(123);

            // Assert
            result.Should().NotBeNull();
            result.Should().Contain("test data");

            mockService.Verify(s => s.GetAsync(123), Times.Once);
        }
    }
}
```

### Running Tests

#### Test Execution Commands
```bash
# Run all tests
dotnet test

# Run tests with detailed output
dotnet test --verbosity detailed

# Run tests in specific project
dotnet test tests/MyLibrary.Tests

# Run tests with filter
dotnet test --filter "FullyQualifiedName~Calculator"
dotnet test --filter "Category=Unit"
dotnet test --filter "Priority=1"

# Run tests with logger
dotnet test --logger trx
dotnet test --logger "console;verbosity=detailed"
dotnet test --logger "html;LogFileName=TestResults.html"

# Parallel execution
dotnet test --parallel
```

#### Code Coverage
```bash
# Install coverlet global tool
dotnet tool install --global coverlet.console

# Run with coverage
dotnet test --collect:"XPlat Code Coverage"
dotnet test /p:CollectCoverage=true /p:CoverletOutputFormat=cobertura

# Generate coverage report
dotnet tool install --global dotnet-reportgenerator-globaltool
reportgenerator -reports:"coverage.cobertura.xml" -targetdir:"coveragereport" -reporttypes:Html
```

#### Test Configuration (runsettings)
```xml
<!-- test.runsettings -->
<?xml version="1.0" encoding="utf-8"?>
<RunSettings>

  <TestRunParameters>
    <Parameter name="ConnectionString" value="Server=localhost;Database=TestDb" />
    <Parameter name="ApiEndpoint" value="https://api.test.com" />
  </TestRunParameters>

  <DataCollectionRunSettings>
    <DataCollectors>
      <DataCollector friendlyName="XPlat code coverage">
        <Configuration>
          <Format>cobertura</Format>
          <Exclude>[*Tests*]*,[*]*.Program</Exclude>
          <ExcludeByAttribute>Obsolete,GeneratedCodeAttribute,CompilerGeneratedAttribute</ExcludeByAttribute>
        </Configuration>
      </DataCollector>
    </DataCollectors>
  </DataCollectionRunSettings>

  <MSTest>
    <Parallelize>
      <Workers>4</Workers>
      <Scope>MethodLevel</Scope>
    </Parallelize>
  </MSTest>

</RunSettings>
```

```bash
# Use runsettings file
dotnet test --settings test.runsettings
```

---

## Code Quality Tools

### Static Analysis

#### Built-in .NET Analyzers
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>

    <!-- Enable all analyzers -->
    <EnableNETAnalyzers>true</EnableNETAnalyzers>
    <AnalysisLevel>latest</AnalysisLevel>
    <AnalysisMode>All</AnalysisMode>

    <!-- Treat warnings as errors -->
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <WarningsAsErrors />
    <WarningsNotAsErrors>CS1591</WarningsNotAsErrors>

    <!-- Code analysis ruleset -->
    <CodeAnalysisRuleSet>$(MSBuildThisFileDirectory)CodeAnalysis.ruleset</CodeAnalysisRuleSet>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.CodeAnalysis.Analyzers" Version="3.3.4" PrivateAssets="all" />
    <PackageReference Include="Microsoft.CodeAnalysis.NetAnalyzers" Version="8.0.0" PrivateAssets="all" />
    <PackageReference Include="StyleCop.Analyzers" Version="1.1.118" PrivateAssets="all" />
    <PackageReference Include="SonarAnalyzer.CSharp" Version="9.12.0.78982" PrivateAssets="all" />
  </ItemGroup>

  <ItemGroup>
    <AdditionalFiles Include="stylecop.json" />
  </ItemGroup>

</Project>
```

#### Custom Ruleset
```xml
<!-- CodeAnalysis.ruleset -->
<?xml version="1.0" encoding="utf-8"?>
<RuleSet Name="Custom Rules" ToolsVersion="16.0">

  <Rules AnalyzerId="Microsoft.CodeAnalysis.CSharp" RuleNamespace="Microsoft.CodeAnalysis.CSharp">
    <Rule Id="CS1591" Action="None" />  <!-- Missing XML comment -->
    <Rule Id="CS0618" Action="Warning" />  <!-- Obsolete member -->
  </Rules>

  <Rules AnalyzerId="Microsoft.CodeAnalysis.CSharp.Features" RuleNamespace="Microsoft.CodeAnalysis.CSharp.Features">
    <Rule Id="IDE0001" Action="Warning" />  <!-- Simplify Names -->
    <Rule Id="IDE0002" Action="Warning" />  <!-- Simplify Member Access -->
  </Rules>

  <Rules AnalyzerId="StyleCop.Analyzers" RuleNamespace="StyleCop.Analyzers">
    <Rule Id="SA1633" Action="None" />  <!-- File header -->
    <Rule Id="SA1200" Action="Error" />  <!-- Using directives placement -->
  </Rules>

</RuleSet>
```

#### StyleCop Configuration
```json
{
  "$schema": "https://raw.githubusercontent.com/DotNetAnalyzers/StyleCopAnalyzers/master/StyleCop.Analyzers/StyleCop.Analyzers/Settings/stylecop.schema.json",
  "settings": {
    "documentationRules": {
      "companyName": "My Company",
      "copyrightText": "Copyright (c) My Company. All rights reserved.",
      "documentExposedElements": true,
      "documentInternalElements": false,
      "documentPrivateElements": false
    },
    "orderingRules": {
      "systemUsingDirectivesFirst": true,
      "usingDirectivesPlacement": "outsideNamespace"
    },
    "namingRules": {
      "allowCommonHungarianPrefixes": true,
      "allowedHungarianPrefixes": ["db", "id"]
    }
  }
}
```

### Code Formatting

#### Using .NET Format
```bash
# Format solution/project
dotnet format
dotnet format --verbosity diagnostic

# Format specific files
dotnet format --include Program.cs

# Format with specific settings
dotnet format --verify-no-changes  # CI mode
dotnet format --fix-style
dotnet format --fix-analyzers
```

#### EditorConfig Integration
```bash
# Format according to .editorconfig
dotnet format --verbosity normal

# Check formatting without fixing
dotnet format --verify-no-changes --verbosity normal
```

### Security Analysis

#### Security Analyzers
```xml
<ItemGroup>
  <PackageReference Include="Microsoft.CodeAnalysis.BannedApiAnalyzers" Version="3.3.4" PrivateAssets="all" />
  <PackageReference Include="Microsoft.CodeAnalysis.PublicApiAnalyzers" Version="3.3.4" PrivateAssets="all" />
  <PackageReference Include="SecurityCodeScan.VS2019" Version="5.6.7" PrivateAssets="all" />
</ItemGroup>

<ItemGroup>
  <AdditionalFiles Include="BannedSymbols.txt" />
  <AdditionalFiles Include="PublicAPI.Shipped.txt" />
  <AdditionalFiles Include="PublicAPI.Unshipped.txt" />
</ItemGroup>
```

#### Banned APIs
```text
<!-- BannedSymbols.txt -->
M:System.Console.WriteLine(System.String); Use ILogger instead
T:System.Web.HttpUtility; Use System.Net.WebUtility instead
M:System.IO.File.ReadAllText(System.String); Use async methods instead
```

---

## Performance Profiling and Diagnostics

### Built-in Diagnostic Tools

#### dotnet-counters
```bash
# Install global tool
dotnet tool install --global dotnet-counters

# List available counters
dotnet-counters list

# Monitor performance counters
dotnet-counters monitor --process-id 1234
dotnet-counters monitor --name MyApp

# Monitor specific counters
dotnet-counters monitor --process-id 1234 --counters System.Runtime[cpu-usage,working-set]

# Collect counters to file
dotnet-counters collect --process-id 1234 --output mycollection
```

#### dotnet-trace
```bash
# Install global tool
dotnet tool install --global dotnet-trace

# Collect trace
dotnet-trace collect --process-id 1234
dotnet-trace collect --name MyApp

# Collect with specific providers
dotnet-trace collect --process-id 1234 --providers Microsoft-DotNETCore-SampleProfiler

# Convert trace format
dotnet-trace convert myapp.nettrace --format chromium
dotnet-trace convert myapp.nettrace --format speedscope
```

#### dotnet-dump
```bash
# Install global tool
dotnet tool install --global dotnet-dump

# Collect dump
dotnet-dump collect --process-id 1234
dotnet-dump collect --name MyApp

# Analyze dump
dotnet-dump analyze mydump.dmp

# Common dump commands
> clrstack
> dumpheap -stat
> gcroot 0x...
> sos DumpHeap -type String
```

### Application Insights Integration

#### Configuration
```csharp
// Program.cs
using Microsoft.ApplicationInsights;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddApplicationInsightsTelemetry();
builder.Services.AddSingleton<TelemetryClient>();

var app = builder.Build();
```

```json
{
  "ApplicationInsights": {
    "ConnectionString": "InstrumentationKey=your-key;IngestionEndpoint=https://your-region.in.applicationinsights.azure.com/"
  },
  "Logging": {
    "ApplicationInsights": {
      "LogLevel": {
        "Default": "Information"
      }
    }
  }
}
```

### Custom Metrics and Logging

#### Performance Monitoring
```csharp
using System.Diagnostics;
using System.Diagnostics.Metrics;

public class PerformanceMonitor
{
    private static readonly Meter Meter = new("MyApp.Performance");
    private static readonly Counter<int> RequestCounter = Meter.CreateCounter<int>("requests_total");
    private static readonly Histogram<double> RequestDuration = Meter.CreateHistogram<double>("request_duration_ms");

    public async Task<T> MeasureAsync<T>(string operationName, Func<Task<T>> operation)
    {
        using var activity = Activity.Current?.Source.StartActivity(operationName);
        var stopwatch = Stopwatch.StartNew();

        try
        {
            RequestCounter.Add(1, new KeyValuePair<string, object?>("operation", operationName));
            var result = await operation();

            activity?.SetStatus(ActivityStatusCode.Ok);
            return result;
        }
        catch (Exception ex)
        {
            activity?.SetStatus(ActivityStatusCode.Error, ex.Message);
            RequestCounter.Add(1,
                new KeyValuePair<string, object?>("operation", operationName),
                new KeyValuePair<string, object?>("result", "error"));
            throw;
        }
        finally
        {
            stopwatch.Stop();
            RequestDuration.Record(stopwatch.Elapsed.TotalMilliseconds,
                new KeyValuePair<string, object?>("operation", operationName));
        }
    }
}
```

---

## Docker Integration

### Dockerfile for .NET Applications

#### Multi-stage Dockerfile
```dockerfile
# Multi-stage build for ASP.NET Core app
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
WORKDIR /app
EXPOSE 8080
EXPOSE 8081

# Create non-root user
RUN adduser --disabled-password --gecos '' appuser && chown -R appuser /app
USER appuser

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project files and restore dependencies
COPY ["src/MyWebApi/MyWebApi.csproj", "src/MyWebApi/"]
COPY ["src/MyLibrary/MyLibrary.csproj", "src/MyLibrary/"]
COPY ["Directory.Build.props", "."]
COPY ["Directory.Packages.props", "."]
COPY ["global.json", "."]

RUN dotnet restore "src/MyWebApi/MyWebApi.csproj"

# Copy source code and build
COPY . .
WORKDIR "/src/src/MyWebApi"
RUN dotnet build "MyWebApi.csproj" -c Release -o /app/build --no-restore

FROM build AS publish
RUN dotnet publish "MyWebApi.csproj" -c Release -o /app/publish --no-restore --no-build

# Final stage
FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["dotnet", "MyWebApi.dll"]
```

#### Optimized Dockerfile for Libraries
```dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Enable BuildKit for better caching
# docker build --progress=plain --no-cache .

# Copy only what's needed for restore
COPY ["*.sln", "."]
COPY ["src/*/*.csproj", "src/"]
COPY ["tests/*/*.csproj", "tests/"]
COPY ["Directory.Build.props", "."]
COPY ["Directory.Packages.props", "."]

# Restore dependencies
RUN dotnet restore

# Copy source and build
COPY . .
RUN dotnet build --configuration Release --no-restore

# Test
RUN dotnet test --configuration Release --no-build --verbosity normal --logger trx --results-directory /testresults

# Pack
RUN dotnet pack --configuration Release --no-build --output /packages

FROM scratch AS packages
COPY --from=build /packages /packages

FROM scratch AS testresults
COPY --from=build /testresults /testresults
```

### Docker Compose for Development

#### docker-compose.yml
```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: src/MyWebApi/Dockerfile
    ports:
      - "5000:8080"
    environment:
      - ASPNETCORE_ENVIRONMENT=Development
      - ConnectionStrings__DefaultConnection=Server=db;Database=MyApp;User Id=sa;Password=YourStrong!Passw0rd;TrustServerCertificate=true
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
    networks:
      - app-network

  db:
    image: mcr.microsoft.com/mssql/server:2022-latest
    environment:
      - ACCEPT_EULA=Y
      - SA_PASSWORD=YourStrong!Passw0rd
    ports:
      - "1433:1433"
    volumes:
      - db_data:/var/opt/mssql
    healthcheck:
      test: /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "$$SA_PASSWORD" -Q "SELECT 1"
      interval: 10s
      timeout: 3s
      retries: 10
      start_period: 10s
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - app-network

volumes:
  db_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

#### docker-compose.override.yml
```yaml
version: '3.8'

services:
  api:
    environment:
      - ASPNETCORE_ENVIRONMENT=Development
      - DOTNET_USE_POLLING_FILE_WATCHER=1
    volumes:
      - ./src:/app/src:ro
      - ./appsettings.Development.json:/app/appsettings.Development.json:ro
    command: ["dotnet", "watch", "run"]
```

### .dockerignore
```gitignore
**/.dockerignore
**/.env
**/.git
**/.gitignore
**/.project
**/.settings
**/.toolstarget
**/.vs
**/.vscode
**/.idea
**/*.*proj.user
**/*.dbmdl
**/*.jfm
**/azds.yaml
**/bin
**/charts
**/docker-compose*
**/Dockerfile*
**/node_modules
**/npm-debug.log
**/obj
**/secrets.dev.yaml
**/values.dev.yaml
LICENSE
README.md
```

---

## Advanced Build Scenarios

### Multi-targeting and Conditional Compilation

#### Multi-targeting Setup
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFrameworks>net8.0;net6.0;net48</TargetFrameworks>
    <LangVersion>latest</LangVersion>
  </PropertyGroup>

  <!-- Framework-specific properties -->
  <PropertyGroup Condition="'$(TargetFramework)' == 'net48'">
    <DefineConstants>$(DefineConstants);NET_FRAMEWORK</DefineConstants>
  </PropertyGroup>

  <PropertyGroup Condition="'$(TargetFramework)' == 'net8.0'">
    <DefineConstants>$(DefineConstants);NET8_0_OR_GREATER</DefineConstants>
  </PropertyGroup>

  <!-- Conditional package references -->
  <ItemGroup Condition="'$(TargetFramework)' == 'net48'">
    <PackageReference Include="System.Memory" Version="4.5.5" />
    <PackageReference Include="Microsoft.Bcl.AsyncInterfaces" Version="8.0.0" />
  </ItemGroup>

  <ItemGroup Condition="'$(TargetFramework)' != 'net48'">
    <PackageReference Include="System.Text.Json" Version="8.0.0" />
  </ItemGroup>

  <!-- Framework-specific files -->
  <ItemGroup Condition="'$(TargetFramework)' == 'net48'">
    <Compile Include="NetFramework\**\*.cs" />
  </ItemGroup>

  <ItemGroup Condition="'$(TargetFramework)' != 'net48'">
    <Compile Include="NetCore\**\*.cs" />
  </ItemGroup>

</Project>
```

#### Conditional Code
```csharp
public class PlatformSpecificCode
{
#if NET8_0_OR_GREATER
    public async Task<string> GetDataAsync()
    {
        using var client = new HttpClient();
        return await client.GetStringAsync("https://api.example.com/data");
    }
#elif NET_FRAMEWORK
    public async Task<string> GetDataAsync()
    {
        using (var client = new HttpClient())
        {
            return await client.GetStringAsync("https://api.example.com/data");
        }
    }
#else
    public Task<string> GetDataAsync()
    {
        throw new NotSupportedException("Platform not supported");
    }
#endif

    public void ProcessJson(string json)
    {
#if NET8_0_OR_GREATER
        var options = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase
        };
        var data = JsonSerializer.Deserialize<MyData>(json, options);
#else
        var data = JsonConvert.DeserializeObject<MyData>(json);
#endif
        // Process data...
    }
}
```

### Build Customization and Automation

#### Custom Build Targets
```xml
<!-- Build/Build.targets -->
<Project>

  <!-- Version from Git -->
  <Target Name="GetVersionFromGit" BeforeTargets="GetAssemblyVersion">
    <Exec Command="git describe --tags --long --always" ConsoleToMSBuild="true">
      <Output TaskParameter="ConsoleOutput" PropertyName="GitVersion" />
    </Exec>
    <PropertyGroup>
      <Version Condition="'$(Version)' == ''">$(GitVersion)</Version>
    </PropertyGroup>
    <Message Text="Version: $(Version)" Importance="high" />
  </Target>

  <!-- Generate build info -->
  <Target Name="GenerateBuildInfo" BeforeTargets="PrepareForBuild">
    <PropertyGroup>
      <BuildInfoFile>$(MSBuildProjectDirectory)\Properties\BuildInfo.cs</BuildInfoFile>
      <BuildTime>$([System.DateTime]::Now.ToString("yyyy-MM-dd HH:mm:ss"))</BuildTime>
      <GitCommit Condition="'$(GitCommit)' == ''">unknown</GitCommit>
    </PropertyGroup>

    <Exec Command="git rev-parse HEAD" ConsoleToMSBuild="true" IgnoreExitCode="true">
      <Output TaskParameter="ConsoleOutput" PropertyName="GitCommit" />
    </Exec>

    <WriteLinesToFile
      File="$(BuildInfoFile)"
      Lines="using System.Reflection%3B

[assembly: AssemblyMetadata(&quot;BuildTime&quot;, &quot;$(BuildTime)&quot;)]
[assembly: AssemblyMetadata(&quot;GitCommit&quot;, &quot;$(GitCommit)&quot;)]
[assembly: AssemblyMetadata(&quot;BuildMachine&quot;, &quot;$(COMPUTERNAME)&quot;)]"
      Overwrite="true" />

    <ItemGroup>
      <Compile Include="$(BuildInfoFile)" />
    </ItemGroup>
  </Target>

  <!-- Clean generated files -->
  <Target Name="CleanGeneratedFiles" BeforeTargets="Clean">
    <Delete Files="$(MSBuildProjectDirectory)\Properties\BuildInfo.cs"
            Condition="Exists('$(MSBuildProjectDirectory)\Properties\BuildInfo.cs')" />
  </Target>

  <!-- Code signing -->
  <Target Name="SignAssemblies" AfterTargets="Build" Condition="'$(SignAssembly)' == 'true'">
    <ItemGroup>
      <AssembliesToSign Include="$(OutputPath)*.dll;$(OutputPath)*.exe" />
    </ItemGroup>

    <Exec Command="signtool sign /f $(CertificateFile) /p $(CertificatePassword) /t http://timestamp.digicert.com @(AssembliesToSign, ' ')"
          Condition="'@(AssembliesToSign)' != ''" />
  </Target>

</Project>
```

### Performance Optimization

#### Incremental and Parallel Builds
```xml
<Project>

  <PropertyGroup>
    <!-- Enable incremental builds -->
    <UseIncrementalBuild>true</UseIncrementalBuild>

    <!-- Parallel builds -->
    <BuildInParallel>true</BuildInParallel>
    <MaxCpuCount>0</MaxCpuCount> <!-- Use all available cores -->

    <!-- Optimize for build speed -->
    <UseSharedCompilation>true</UseSharedCompilation>
    <ProduceReferenceAssembly>true</ProduceReferenceAssembly>
  </PropertyGroup>

</Project>
```

```bash
# Command line parallel builds
dotnet build --maxcpucount:0  # Use all cores
dotnet build -m:4             # Use 4 cores
dotnet build --no-dependencies # Skip dependency builds
```

#### Build Caching
```xml
<Project>

  <PropertyGroup>
    <!-- Enable deterministic builds -->
    <Deterministic>true</Deterministic>
    <PathMap>$(MSBuildProjectDirectory)=.</PathMap>

    <!-- Output caching -->
    <CacheOutputDirectory>$(MSBuildThisFileDirectory).cache</CacheOutputDirectory>
  </PropertyGroup>

  <!-- Custom caching target -->
  <Target Name="CacheRestorePackages"
          BeforeTargets="RestorePackages"
          Condition="!Exists('$(CacheOutputDirectory)\packages.cache')">
    <MakeDir Directories="$(CacheOutputDirectory)" />
    <Touch Files="$(CacheOutputDirectory)\packages.cache" AlwaysCreate="true" />
  </Target>

</Project>
```

### CI/CD Integration

#### GitHub Actions Workflow
```yaml
# .github/workflows/build.yml
name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  DOTNET_VERSION: '8.0.x'
  DOTNET_SKIP_FIRST_TIME_EXPERIENCE: 1
  DOTNET_NOLOGO: true

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Needed for GitVersion

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ env.DOTNET_VERSION }}

    - name: Cache NuGet packages
      uses: actions/cache@v3
      with:
        path: ~/.nuget/packages
        key: ${{ runner.os }}-nuget-${{ hashFiles('**/*.csproj', '**/Directory.Packages.props') }}
        restore-keys: ${{ runner.os }}-nuget-

    - name: Restore dependencies
      run: dotnet restore --locked-mode

    - name: Build
      run: dotnet build --no-restore --configuration Release

    - name: Test
      run: dotnet test --no-build --configuration Release --verbosity normal --collect:"XPlat Code Coverage" --results-directory TestResults

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        directory: TestResults
        fail_ci_if_error: true

    - name: Pack
      run: dotnet pack --no-build --configuration Release --output nupkgs

    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: packages
        path: nupkgs/*.nupkg
```

This comprehensive guide covers the essential aspects of C# local development and build systems. Each section provides practical examples that can be directly applied to real-world projects, from basic project setup to advanced CI/CD scenarios. The examples demonstrate best practices for maintainable, scalable, and efficient .NET development workflows.