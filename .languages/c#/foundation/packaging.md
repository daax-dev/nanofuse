# C# Package Management & NuGet

## Table of Contents
1. [NuGet Fundamentals](#nuget-fundamentals)
2. [Creating NuGet Packages](#creating-nuget-packages)
3. [Package Versioning](#package-versioning)
4. [Package Metadata](#package-metadata)
5. [Dependency Management](#dependency-management)
6. [Local Repositories](#local-repositories)
7. [Security & Signing](#security--signing)
8. [Multi-Targeting](#multi-targeting)
9. [Advanced Scenarios](#advanced-scenarios)
10. [Best Practices](#best-practices)
11. [Private Repositories](#private-repositories)
12. [Analysis & Tooling](#analysis--tooling)

## NuGet Fundamentals

### Overview
NuGet is the package manager for .NET, providing a centralized way to share and consume libraries, tools, and other components.

### Core Concepts

#### Package Structure
```
MyPackage.1.2.3.nupkg
├── lib/
│   ├── net6.0/
│   │   └── MyLibrary.dll
│   └── netstandard2.0/
│       └── MyLibrary.dll
├── content/
├── tools/
├── build/
│   └── MyPackage.targets
└── MyPackage.nuspec
```

#### Package Types
```xml
<!-- Library Package -->
<PackageType>Dependency</PackageType>

<!-- Tool Package -->
<PackageType>Tool</PackageType>

<!-- Template Package -->
<PackageType>Template</PackageType>

<!-- MSBuild SDK -->
<PackageType>MSBuildSdk</PackageType>
```

### NuGet CLI Commands

#### Essential Commands
```bash
# Install globally
dotnet tool install -g nuget

# Package operations
nuget pack MyProject.csproj
nuget push MyPackage.1.0.0.nupkg -Source https://api.nuget.org/v3/index.json
nuget delete MyPackage 1.0.0 -Source https://api.nuget.org/v3/index.json

# Local operations
nuget add MyPackage.1.0.0.nupkg -source ./local-packages
nuget sources add -name "Local" -source ./local-packages

# Verification
nuget verify -Signatures MyPackage.1.0.0.nupkg
```

#### Package Management
```bash
# .NET CLI package commands
dotnet add package Newtonsoft.Json
dotnet add package Microsoft.Extensions.Logging --version 6.0.0
dotnet remove package Newtonsoft.Json
dotnet list package
dotnet list package --outdated
dotnet restore
```

## Creating NuGet Packages

### Project-Based Packaging

#### Basic Package Configuration
```xml
<!-- MyLibrary.csproj -->
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>

    <!-- Package Information -->
    <PackageId>MyCompany.MyLibrary</PackageId>
    <Version>1.2.3</Version>
    <Authors>John Doe</Authors>
    <Company>MyCompany</Company>
    <Product>MyLibrary</Product>
    <Description>A comprehensive library for handling data operations</Description>
    <Copyright>Copyright © MyCompany 2024</Copyright>

    <!-- Package Metadata -->
    <PackageTags>data;orm;entity;framework</PackageTags>
    <PackageProjectUrl>https://github.com/mycompany/mylibrary</PackageProjectUrl>
    <RepositoryUrl>https://github.com/mycompany/mylibrary.git</RepositoryUrl>
    <RepositoryType>git</RepositoryType>
    <PackageLicenseExpression>MIT</PackageLicenseExpression>
    <PackageIcon>icon.png</PackageIcon>
    <PackageReadmeFile>README.md</PackageReadmeFile>

    <!-- Build Configuration -->
    <GeneratePackageOnBuild>true</GeneratePackageOnBuild>
    <IncludeSymbols>true</IncludeSymbols>
    <SymbolPackageFormat>snupkg</SymbolPackageFormat>
    <GenerateDocumentationFile>true</GenerateDocumentationFile>
  </PropertyGroup>

  <ItemGroup>
    <None Include="icon.png" Pack="true" PackagePath="\" />
    <None Include="README.md" Pack="true" PackagePath="\" />
  </ItemGroup>

  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.3" />
    <PackageReference Include="Microsoft.Extensions.Logging" Version="6.0.0" />
  </ItemGroup>

</Project>
```

#### Advanced Package Properties
```xml
<PropertyGroup>
  <!-- Release Notes -->
  <PackageReleaseNotes>
    Version 1.2.3:
    - Added async support for data operations
    - Fixed memory leak in connection pooling
    - Breaking change: Renamed DataContext to DataProvider
  </PackageReleaseNotes>

  <!-- Requirements -->
  <PackageRequireLicenseAcceptance>true</PackageRequireLicenseAcceptance>
  <DevelopmentDependency>true</DevelopmentDependency>

  <!-- MSBuild Integration -->
  <IsTool>true</IsTool>
  <ToolCommandName>mytool</ToolCommandName>

  <!-- Content Files -->
  <ContentTargetFolders>content</ContentTargetFolders>
  <BuildAction>Content</BuildAction>

  <!-- Package Validation -->
  <EnablePackageValidation>true</EnablePackageValidation>
  <PackageValidationBaselineVersion>1.0.0</PackageValidationBaselineVersion>
</PropertyGroup>
```

### NuSpec-Based Packaging

#### Complete NuSpec Example
```xml
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
  <metadata>
    <id>MyCompany.MyLibrary</id>
    <version>1.2.3</version>
    <title>MyLibrary - Data Operations</title>
    <authors>John Doe, Jane Smith</authors>
    <owners>MyCompany</owners>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
    <license type="expression">MIT</license>
    <licenseUrl>https://licenses.nuget.org/MIT</licenseUrl>
    <icon>images\icon.png</icon>
    <projectUrl>https://github.com/mycompany/mylibrary</projectUrl>
    <description>A comprehensive library for handling data operations with support for multiple databases and async operations.</description>
    <releaseNotes>
      Version 1.2.3:
      - Added async support for data operations
      - Fixed memory leak in connection pooling
      - Breaking change: Renamed DataContext to DataProvider
    </releaseNotes>
    <copyright>Copyright © MyCompany 2024</copyright>
    <tags>data orm entity framework async</tags>
    <repository type="git" url="https://github.com/mycompany/mylibrary.git" />

    <dependencies>
      <group targetFramework=".NETStandard2.0">
        <dependency id="Newtonsoft.Json" version="13.0.3" />
        <dependency id="Microsoft.Extensions.Logging" version="[6.0.0,7.0.0)" />
      </group>
      <group targetFramework="net6.0">
        <dependency id="Newtonsoft.Json" version="13.0.3" />
        <dependency id="Microsoft.Extensions.Logging" version="[6.0.0,8.0.0)" />
      </group>
    </dependencies>

    <frameworkAssemblies>
      <frameworkAssembly assemblyName="System.Data" targetFramework="net472" />
    </frameworkAssemblies>

    <contentFiles>
      <files include="cs/netstandard2.0/config/appsettings.json" buildAction="Content" copyToOutput="true" />
    </contentFiles>
  </metadata>

  <files>
    <!-- Library files -->
    <file src="bin\Release\netstandard2.0\MyLibrary.dll" target="lib\netstandard2.0" />
    <file src="bin\Release\net6.0\MyLibrary.dll" target="lib\net6.0" />

    <!-- Symbol files -->
    <file src="bin\Release\netstandard2.0\MyLibrary.pdb" target="lib\netstandard2.0" />
    <file src="bin\Release\net6.0\MyLibrary.pdb" target="lib\net6.0" />

    <!-- Documentation -->
    <file src="bin\Release\netstandard2.0\MyLibrary.xml" target="lib\netstandard2.0" />
    <file src="bin\Release\net6.0\MyLibrary.xml" target="lib\net6.0" />

    <!-- Content files -->
    <file src="content\**" target="content" />
    <file src="contentFiles\**" target="contentFiles" />

    <!-- Build files -->
    <file src="build\MyLibrary.targets" target="build" />
    <file src="build\MyLibrary.props" target="build" />

    <!-- Tools -->
    <file src="tools\**" target="tools" exclude="**\*.pdb" />

    <!-- Icon and readme -->
    <file src="icon.png" target="images\" />
    <file src="README.md" target="" />
  </files>
</package>
```

### Build Integration

#### MSBuild Targets
```xml
<!-- MyLibrary.targets -->
<Project>
  <PropertyGroup>
    <MyLibraryVersion>1.2.3</MyLibraryVersion>
  </PropertyGroup>

  <Target Name="MyLibrarySetup" BeforeTargets="Build">
    <Message Text="Setting up MyLibrary $(MyLibraryVersion)" Importance="high" />
    <ItemGroup>
      <Reference Include="$(MSBuildThisFileDirectory)..\lib\$(TargetFramework)\MyLibrary.dll" />
    </ItemGroup>
  </Target>

  <Target Name="MyLibraryPostBuild" AfterTargets="Build">
    <Copy SourceFiles="$(MSBuildThisFileDirectory)..\content\config.json"
          DestinationFolder="$(OutputPath)" />
  </Target>
</Project>
```

#### Props File
```xml
<!-- MyLibrary.props -->
<Project>
  <PropertyGroup>
    <DefineConstants>$(DefineConstants);MYLIBRARY_AVAILABLE</DefineConstants>
    <MyLibraryConfigPath>$(MSBuildProjectDirectory)\mylibrary.config</MyLibraryConfigPath>
  </PropertyGroup>

  <ItemGroup>
    <Analyzer Include="$(MSBuildThisFileDirectory)..\analyzers\MyLibrary.Analyzers.dll" />
  </ItemGroup>
</Project>
```

## Package Versioning

### Semantic Versioning (SemVer)

#### Version Format
```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]

Examples:
1.0.0           - Initial release
1.2.3           - Patch release
2.0.0           - Major release (breaking changes)
1.3.0-alpha.1   - Pre-release
1.3.0-beta.2    - Beta release
1.3.0-rc.1      - Release candidate
1.3.0+20240115  - Build metadata
```

#### Version Increment Rules
```csharp
// MAJOR: Breaking changes
public class DataContext  // v1.x
{
    public void SaveChanges() { }
}

public class DataProvider // v2.0 - Breaking change (renamed class)
{
    public Task SaveChangesAsync() { } // Breaking change (method signature)
}

// MINOR: New features (backward compatible)
public class DataProvider // v2.1 - Added new method
{
    public Task SaveChangesAsync() { }
    public void SaveChanges() { } // Added backward compatibility
}

// PATCH: Bug fixes
public class DataProvider // v2.1.1 - Fixed bug in SaveChanges
{
    public Task SaveChangesAsync() { }
    public void SaveChanges()
    {
        // Fixed null reference exception
        if (context != null)
        {
            context.SaveChanges();
        }
    }
}
```

### Version Management

#### Assembly Versioning
```xml
<PropertyGroup>
  <!-- File version (display) -->
  <FileVersion>1.2.3.456</FileVersion>

  <!-- Assembly version (binary compatibility) -->
  <AssemblyVersion>1.0.0.0</AssemblyVersion>

  <!-- Informational version (product) -->
  <InformationalVersion>1.2.3-beta.1+sha.1234567</InformationalVersion>

  <!-- Package version -->
  <Version>1.2.3-beta.1</Version>
</PropertyGroup>
```

#### Automated Versioning
```xml
<!-- Using GitVersion -->
<PropertyGroup>
  <Version>$(GitVersion_NuGetVersion)</Version>
  <AssemblyVersion>$(GitVersion_AssemblySemVer)</AssemblyVersion>
  <FileVersion>$(GitVersion_AssemblySemFileVer)</FileVersion>
  <InformationalVersion>$(GitVersion_InformationalVersion)</InformationalVersion>
</PropertyGroup>

<ItemGroup>
  <PackageReference Include="GitVersion.MsBuild" Version="5.12.0">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

#### Version Ranges
```xml
<!-- Dependency version ranges -->
<PackageReference Include="Newtonsoft.Json" Version="13.0.3" />              <!-- Exact -->
<PackageReference Include="Microsoft.Extensions.Logging" Version="[6.0.0]" /> <!-- Exact (brackets) -->
<PackageReference Include="System.Text.Json" Version="[6.0.0,7.0.0)" />      <!-- Range (min inclusive, max exclusive) -->
<PackageReference Include="AutoMapper" Version="[12.0.0,)" />                  <!-- Minimum version -->
<PackageReference Include="FluentValidation" Version="(,11.0.0)" />           <!-- Maximum version -->
<PackageReference Include="Serilog" Version="*" />                            <!-- Latest (not recommended) -->
```

## Package Metadata

### Essential Metadata

#### Basic Information
```xml
<PropertyGroup>
  <!-- Identity -->
  <PackageId>MyCompany.MyLibrary</PackageId>
  <Version>1.2.3</Version>
  <Title>MyLibrary - Data Operations Framework</Title>

  <!-- Ownership -->
  <Authors>John Doe, Jane Smith</Authors>
  <Company>MyCompany Inc.</Company>
  <Owners>MyCompany</Owners>

  <!-- Description -->
  <Description>
    A comprehensive data operations framework providing:
    - Entity Framework integration
    - Async/await support
    - Connection pooling
    - Transaction management
    - Caching mechanisms
  </Description>
  <Summary>High-performance data operations framework for .NET</Summary>

  <!-- Classification -->
  <PackageTags>data;orm;entity;framework;async;database;sql</PackageTags>
  <Copyright>Copyright © MyCompany Inc. 2024</Copyright>
</PropertyGroup>
```

#### Documentation & Links
```xml
<PropertyGroup>
  <!-- Project Information -->
  <PackageProjectUrl>https://github.com/mycompany/mylibrary</PackageProjectUrl>
  <RepositoryUrl>https://github.com/mycompany/mylibrary.git</RepositoryUrl>
  <RepositoryType>git</RepositoryType>
  <RepositoryBranch>main</RepositoryBranch>
  <RepositoryCommit>abc123def456</RepositoryCommit>

  <!-- Documentation -->
  <PackageReadmeFile>README.md</PackageReadmeFile>
  <PackageIcon>icon.png</PackageIcon>
  <PackageIconUrl>https://raw.githubusercontent.com/mycompany/mylibrary/main/icon.png</PackageIconUrl>

  <!-- Release Information -->
  <PackageReleaseNotes>docs/CHANGELOG.md</PackageReleaseNotes>
  <PackageReleaseNotes>
    ## Version 1.2.3

    ### Features
    - Added support for async operations
    - Implemented connection pooling

    ### Bug Fixes
    - Fixed memory leak in data context
    - Resolved transaction deadlock issues

    ### Breaking Changes
    - Renamed `DataContext` to `DataProvider`
    - Changed method signatures for async support
  </PackageReleaseNotes>
</PropertyGroup>
```

### License Management

#### License Expression
```xml
<PropertyGroup>
  <!-- SPDX License Expression -->
  <PackageLicenseExpression>MIT</PackageLicenseExpression>
  <PackageLicenseExpression>Apache-2.0</PackageLicenseExpression>
  <PackageLicenseExpression>GPL-3.0-or-later</PackageLicenseExpression>
  <PackageLicenseExpression>MIT OR Apache-2.0</PackageLicenseExpression>
  <PackageLicenseExpression>MIT AND Apache-2.0</PackageLicenseExpression>
</PropertyGroup>
```

#### License File
```xml
<PropertyGroup>
  <PackageLicenseFile>LICENSE.txt</PackageLicenseFile>
  <PackageRequireLicenseAcceptance>true</PackageRequireLicenseAcceptance>
</PropertyGroup>

<ItemGroup>
  <None Include="LICENSE.txt" Pack="true" PackagePath="" />
</ItemGroup>
```

### Content Files

#### Content Configuration
```xml
<ItemGroup>
  <!-- Static content -->
  <Content Include="templates\**\*" PackagePath="content\templates" />

  <!-- Build-time content -->
  <None Include="config\appsettings.template.json" Pack="true" PackagePath="content\config" />

  <!-- Content files (new format) -->
  <None Include="contentFiles\cs\any\Config\Settings.cs.pp" Pack="true" PackagePath="contentFiles\cs\any\Config\" />
</ItemGroup>

<!-- Content file metadata -->
<PropertyGroup>
  <ContentTargetFolders>content;contentFiles</ContentTargetFolders>
</PropertyGroup>
```

#### Parameter Replacement
```csharp
// Settings.cs.pp (parameter replacement template)
namespace $rootnamespace$.Config
{
    public static class Settings
    {
        public const string Version = "$version$";
        public const string PackageId = "$id$";
        public const string Author = "$author$";
    }
}
```

## Dependency Management

### Dependency Declaration

#### Package References
```xml
<ItemGroup>
  <!-- Production dependencies -->
  <PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="6.0.0" />
  <PackageReference Include="Microsoft.Extensions.Logging" Version="6.0.0" />
  <PackageReference Include="Newtonsoft.Json" Version="13.0.3" />

  <!-- Development dependencies -->
  <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.5.0">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
  </PackageReference>

  <!-- Analyzer dependencies -->
  <PackageReference Include="StyleCop.Analyzers" Version="1.1.118">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>analyzers</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

#### Framework References
```xml
<ItemGroup>
  <!-- .NET Framework assemblies -->
  <Reference Include="System.Data" Condition="'$(TargetFramework)' == 'net472'" />
  <Reference Include="System.Configuration" Condition="'$(TargetFramework)' == 'net472'" />

  <!-- Framework references (modern) -->
  <FrameworkReference Include="Microsoft.AspNetCore.App" Condition="'$(TargetFramework)' == 'net6.0'" />
</ItemGroup>
```

### Dependency Resolution

#### Package Lock Files
```json
// packages.lock.json
{
  "version": 1,
  "dependencies": {
    "net6.0": {
      "Microsoft.Extensions.Logging": {
        "type": "Direct",
        "requested": "[6.0.0, )",
        "resolved": "6.0.0",
        "contentHash": "eIbyj40QDg1NDz0HBW0S5f3wrLVnKWnDJ/JtZ+yJDFnDQC5MrvdlIXfFoV8lqmPE5rz8kP45L03J4e9rMr7p3A=="
      },
      "Microsoft.Extensions.Logging.Abstractions": {
        "type": "Transitive",
        "resolved": "6.0.0",
        "contentHash": "0M3y7WswOV/mE4psANi9h0CU69mZY58sA8CEY1/Bvv+qWfpJV2XQEa3nDAUt4/A2/yk7GJdQ+hBfAUbFEjNAOTg=="
      }
    }
  }
}
```

#### Central Package Management
```xml
<!-- Directory.Packages.props -->
<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
    <CentralPackageTransitivePinningEnabled>true</CentralPackageTransitivePinningEnabled>
  </PropertyGroup>

  <ItemGroup>
    <PackageVersion Include="Microsoft.Extensions.Logging" Version="6.0.0" />
    <PackageVersion Include="Microsoft.Extensions.DependencyInjection" Version="6.0.0" />
    <PackageVersion Include="Newtonsoft.Json" Version="13.0.3" />
    <PackageVersion Include="xunit" Version="2.4.2" />
    <PackageVersion Include="xunit.runner.visualstudio" Version="2.4.3" />
  </ItemGroup>

  <!-- Global package references -->
  <ItemGroup>
    <GlobalPackageReference Include="Microsoft.SourceLink.GitHub" Version="1.1.1" />
  </ItemGroup>
</Project>
```

### Package Sources

#### NuGet Configuration
```xml
<!-- nuget.config -->
<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
    <add key="myget" value="https://www.myget.org/F/mycompany/api/v3/index.json" />
    <add key="local" value="./local-packages" />
    <add key="azure-artifacts" value="https://pkgs.dev.azure.com/myorg/_packaging/mycompany/nuget/v3/index.json" />
  </packageSources>

  <packageSourceCredentials>
    <azure-artifacts>
      <add key="Username" value="anything" />
      <add key="ClearTextPassword" value="%AZURE_ARTIFACTS_PAT%" />
    </azure-artifacts>
  </packageSourceCredentials>

  <packageRestore>
    <add key="enabled" value="True" />
    <add key="automatic" value="True" />
  </packageRestore>

  <bindingRedirects>
    <add key="skip" value="False" />
  </bindingRedirects>

  <packageManagement>
    <add key="format" value="1" />
    <add key="disabled" value="False" />
  </packageManagement>
</configuration>
```

## Local Repositories

### Local Package Development

#### Local Source Setup
```bash
# Create local package source
mkdir ./local-packages
nuget sources add -name "Local" -source "./local-packages"

# Add package to local source
nuget add MyPackage.1.0.0.nupkg -source ./local-packages

# List local packages
nuget list -source ./local-packages
```

#### Project-to-Project References
```xml
<!-- During development -->
<ItemGroup>
  <ProjectReference Include="..\MyLibrary\MyLibrary.csproj" />
</ItemGroup>

<!-- For packaging -->
<ItemGroup>
  <PackageReference Include="MyCompany.MyLibrary" Version="1.2.3" />
</ItemGroup>
```

### Development Workflows

#### Symbol Packages
```xml
<PropertyGroup>
  <!-- Generate symbol package -->
  <IncludeSymbols>true</IncludeSymbols>
  <SymbolPackageFormat>snupkg</SymbolPackageFormat>

  <!-- Source link -->
  <PublishRepositoryUrl>true</PublishRepositoryUrl>
  <EmbedUntrackedSources>true</EmbedUntrackedSources>
  <DebugType>embedded</DebugType>
</PropertyGroup>

<ItemGroup>
  <PackageReference Include="Microsoft.SourceLink.GitHub" Version="1.1.1">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

#### Local Testing
```bash
# Pack locally
dotnet pack -c Release -o ./local-packages

# Test installation
cd TestProject
dotnet add package MyCompany.MyLibrary --source ../local-packages --version 1.2.3-local

# Clear cache if needed
dotnet nuget locals all --clear
```

## Security & Signing

### Package Signing

#### Code Signing Certificate
```bash
# Sign package with certificate
nuget sign MyPackage.1.0.0.nupkg -CertificateFilePath certificate.p12 -CertificatePassword "password" -Timestamper http://timestamp.digicert.com

# Sign with certificate store
nuget sign MyPackage.1.0.0.nupkg -CertificateSubjectName "My Company" -Timestamper http://timestamp.digicert.com

# Verify signature
nuget verify -Signatures MyPackage.1.0.0.nupkg
```

#### Repository Signing
```xml
<!-- Project configuration for repository signing -->
<PropertyGroup>
  <SignPackage>true</SignPackage>
  <PackageCertificateThumbprint>1234567890ABCDEF1234567890ABCDEF12345678</PackageCertificateThumbprint>
  <PackageSigningEnabled>true</PackageSigningEnabled>
  <RepositorySigningEnabled>true</RepositorySigningEnabled>
</PropertyGroup>
```

### Security Scanning

#### Vulnerability Assessment
```bash
# Check for vulnerabilities
dotnet list package --vulnerable

# Check for deprecated packages
dotnet list package --deprecated

# Audit packages
dotnet restore --packages packages
dotnet list package --vulnerable --include-transitive

# Generate security report
dotnet list package --vulnerable --format json > vulnerability-report.json
```

#### Package Validation
```xml
<PropertyGroup>
  <!-- Enable package validation -->
  <EnablePackageValidation>true</EnablePackageValidation>
  <PackageValidationBaselineVersion>1.0.0</PackageValidationBaselineVersion>

  <!-- Generate compatibility reports -->
  <GenerateCompatibilitySuppressionFile>true</GenerateCompatibilitySuppressionFile>
  <CompatibilitySuppressionFilePath>CompatibilitySuppressions.xml</CompatibilitySuppressionFilePath>
</PropertyGroup>

<ItemGroup>
  <PackageReference Include="Microsoft.DotNet.PackageValidation" Version="1.0.0-preview.7.21379.12">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>build</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

### Trust Policies

#### NuGet Trust Configuration
```xml
<!-- nuget.config -->
<configuration>
  <trustedSigners>
    <author name="Microsoft">
      <certificate fingerprint="3F9001EA83C560D712C24CF213C3D312CB3BFF51EE89435D3430BD06B5D0EECE"
                   hashAlgorithm="SHA256"
                   allowUntrustedRoot="false" />
    </author>

    <repository name="nuget.org" serviceIndex="https://api.nuget.org/v3/index.json">
      <certificate fingerprint="0E5F38F57DC1BCC806D8494F4F90FBCEDD988B46760709CBEEC6F4219AA6157D"
                   hashAlgorithm="SHA256"
                   allowUntrustedRoot="false" />
    </repository>
  </trustedSigners>

  <packageSourceMapping>
    <packageSource key="nuget.org">
      <package pattern="Microsoft.*" />
      <package pattern="System.*" />
      <package pattern="NETStandard.*" />
    </packageSource>
    <packageSource key="contoso">
      <package pattern="Contoso.*" />
    </packageSource>
  </packageSourceMapping>
</configuration>
```

## Multi-Targeting

### Target Framework Support

#### Multi-Target Configuration
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <!-- Multiple target frameworks -->
    <TargetFrameworks>net472;netstandard2.0;netstandard2.1;net6.0;net7.0;net8.0</TargetFrameworks>

    <!-- Framework-specific properties -->
    <LangVersion Condition="'$(TargetFramework)' == 'net472'">7.3</LangVersion>
    <LangVersion Condition="'$(TargetFramework)' != 'net472'">latest</LangVersion>

    <!-- Conditional compilation symbols -->
    <DefineConstants Condition="'$(TargetFramework)' == 'net472'">$(DefineConstants);NET472</DefineConstants>
    <DefineConstants Condition="'$(TargetFramework)' == 'netstandard2.0'">$(DefineConstants);NETSTANDARD2_0</DefineConstants>
  </PropertyGroup>

  <!-- Framework-specific dependencies -->
  <ItemGroup Condition="'$(TargetFramework)' == 'net472'">
    <Reference Include="System.Net.Http" />
    <PackageReference Include="Microsoft.Bcl.AsyncInterfaces" Version="7.0.0" />
  </ItemGroup>

  <ItemGroup Condition="'$(TargetFramework)' == 'netstandard2.0'">
    <PackageReference Include="Microsoft.Bcl.AsyncInterfaces" Version="7.0.0" />
    <PackageReference Include="System.Text.Json" Version="7.0.3" />
  </ItemGroup>

  <ItemGroup Condition="'$(TargetFramework)' == 'net6.0' OR '$(TargetFramework)' == 'net7.0' OR '$(TargetFramework)' == 'net8.0'">
    <!-- Modern .NET has built-in support -->
  </ItemGroup>
</Project>
```

#### Conditional Compilation
```csharp
using System;
using System.Threading.Tasks;
#if NET472
using System.Net.Http;
#endif
#if NETSTANDARD2_0 || NET472
using Microsoft.Bcl.AsyncInterfaces;
#endif

namespace MyLibrary
{
    public class HttpService
    {
#if NET6_0_OR_GREATER
        private static readonly HttpClient httpClient = new();

        public async Task<string> GetDataAsync(string url)
        {
            return await httpClient.GetStringAsync(url);
        }
#elif NETSTANDARD2_0
        private static readonly HttpClient httpClient = new HttpClient();

        public async Task<string> GetDataAsync(string url)
        {
            using var response = await httpClient.GetAsync(url);
            return await response.Content.ReadAsStringAsync();
        }
#elif NET472
        private static readonly HttpClient httpClient = new HttpClient();

        public Task<string> GetDataAsync(string url)
        {
            return httpClient.GetStringAsync(url);
        }
#endif

#if NET5_0_OR_GREATER
        public async IAsyncEnumerable<string> GetStreamAsync(string url)
        {
            using var stream = await httpClient.GetStreamAsync(url);
            using var reader = new StreamReader(stream);

            string line;
            while ((line = await reader.ReadLineAsync()) != null)
            {
                yield return line;
            }
        }
#endif
    }
}
```

### Platform-Specific Packages

#### Runtime Identifiers
```xml
<PropertyGroup>
  <!-- Multi-platform support -->
  <RuntimeIdentifiers>win-x64;win-x86;linux-x64;osx-x64;osx-arm64</RuntimeIdentifiers>

  <!-- Platform-specific assets -->
  <PackageId>MyCompany.MyLibrary</PackageId>
  <PackageId Condition="'$(RuntimeIdentifier)' == 'win-x64'">MyCompany.MyLibrary.Windows</PackageId>
  <PackageId Condition="'$(RuntimeIdentifier)' == 'linux-x64'">MyCompany.MyLibrary.Linux</PackageId>
</PropertyGroup>

<!-- Native dependencies -->
<ItemGroup Condition="'$(RuntimeIdentifier)' == 'win-x64'">
  <Content Include="runtimes\win-x64\native\mylibrary.dll" PackagePath="runtimes\win-x64\native\" />
</ItemGroup>

<ItemGroup Condition="'$(RuntimeIdentifier)' == 'linux-x64'">
  <Content Include="runtimes\linux-x64\native\libmylibrary.so" PackagePath="runtimes\linux-x64\native\" />
</ItemGroup>
```

## Advanced Scenarios

### Source Generators

#### Source Generator Package
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>netstandard2.0</TargetFramework>
    <PackageId>MyCompany.SourceGenerators</PackageId>
    <IncludeBuildOutput>false</IncludeBuildOutput>
    <GeneratePackageOnBuild>true</GeneratePackageOnBuild>
    <DevelopmentDependency>true</DevelopmentDependency>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.CodeAnalysis.CSharp" Version="4.0.1" PrivateAssets="all" />
    <PackageReference Include="Microsoft.CodeAnalysis.Analyzers" Version="3.3.4" PrivateAssets="all" />
  </ItemGroup>

  <ItemGroup>
    <Analyzer Include="$(OutputPath)\MyCompany.SourceGenerators.dll" />
  </ItemGroup>

  <!-- Package the source generator -->
  <ItemGroup>
    <None Include="$(OutputPath)\MyCompany.SourceGenerators.dll" Pack="true" PackagePath="analyzers\dotnet\cs\MyCompany.SourceGenerators.dll" Visible="false" />
  </ItemGroup>
</Project>
```

#### Source Generator Implementation
```csharp
[Generator]
public class EntityGenerator : ISourceGenerator
{
    public void Initialize(GeneratorInitializationContext context)
    {
        context.RegisterForSyntaxNotifications(() => new SyntaxReceiver());
    }

    public void Execute(GeneratorExecutionContext context)
    {
        var receiver = (SyntaxReceiver)context.SyntaxReceiver;

        foreach (var classDeclaration in receiver.CandidateClasses)
        {
            var model = context.Compilation.GetSemanticModel(classDeclaration.SyntaxTree);
            var symbol = model.GetDeclaredSymbol(classDeclaration);

            if (HasEntityAttribute(symbol))
            {
                var source = GenerateEntityExtensions(symbol);
                context.AddSource($"{symbol.Name}Extensions.g.cs", source);
            }
        }
    }
}
```

### Analyzers & Code Fixes

#### Analyzer Package Configuration
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>netstandard2.0</TargetFramework>
    <PackageId>MyCompany.Analyzers</PackageId>
    <PackageVersion>1.0.0</PackageVersion>
    <IncludeBuildOutput>false</IncludeBuildOutput>
    <DevelopmentDependency>true</DevelopmentDependency>
    <NoPackageAnalysis>true</NoPackageAnalysis>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.CodeAnalysis.CSharp" Version="4.0.1" PrivateAssets="all" />
    <PackageReference Include="Microsoft.CodeAnalysis.Analyzers" Version="3.3.4" PrivateAssets="all" />
  </ItemGroup>

  <ItemGroup>
    <None Include="tools\install.ps1" Pack="true" PackagePath="tools\install.ps1" />
    <None Include="tools\uninstall.ps1" Pack="true" PackagePath="tools\uninstall.ps1" />
  </ItemGroup>

  <ItemGroup>
    <Analyzer Include="$(OutputPath)\MyCompany.Analyzers.dll" />
  </ItemGroup>
</Project>
```

### MSBuild SDK

#### Custom SDK Structure
```
MyCompany.Build.Sdk/
├── Sdk/
│   ├── Sdk.props
│   ├── Sdk.targets
│   └── tools/
├── build/
│   ├── MyCompany.Build.Sdk.props
│   └── MyCompany.Build.Sdk.targets
└── MyCompany.Build.Sdk.csproj
```

#### SDK Implementation
```xml
<!-- Sdk/Sdk.props -->
<Project>
  <PropertyGroup>
    <MSBuildAllProjects>$(MSBuildAllProjects);$(MSBuildThisFileFullPath)</MSBuildAllProjects>
    <DefaultItemExcludes>$(DefaultItemExcludes);bin\**;obj\**;**\*.user;**\*.vspscc</DefaultItemExcludes>
  </PropertyGroup>

  <!-- Import standard .NET SDK -->
  <Import Project="Sdk.props" Sdk="Microsoft.NET.Sdk" />

  <!-- Custom properties -->
  <PropertyGroup>
    <MyCompanyVersion>1.0.0</MyCompanyVersion>
    <GenerateAssemblyInfo>true</GenerateAssemblyInfo>
    <GenerateDocumentationFile>true</GenerateDocumentationFile>
  </PropertyGroup>
</Project>
```

## Best Practices

### Package Design

#### API Design Guidelines
```csharp
// Good: Clear, consistent naming
public class UserService
{
    public Task<User> GetUserAsync(int id) { }
    public Task<User> CreateUserAsync(User user) { }
    public Task<User> UpdateUserAsync(int id, User user) { }
    public Task DeleteUserAsync(int id) { }
}

// Good: Async overloads
public class FileProcessor
{
    public void ProcessFile(string path) { }
    public Task ProcessFileAsync(string path) { }
    public Task ProcessFileAsync(string path, CancellationToken cancellationToken) { }
}

// Good: Configuration options
public class DatabaseOptions
{
    public string ConnectionString { get; set; }
    public int CommandTimeout { get; set; } = 30;
    public bool EnableRetryOnFailure { get; set; } = true;
    public int MaxRetryCount { get; set; } = 3;
}

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddMyDatabase(
        this IServiceCollection services,
        Action<DatabaseOptions> configure = null)
    {
        services.Configure<DatabaseOptions>(configure ?? (_ => { }));
        services.AddScoped<IDatabaseService, DatabaseService>();
        return services;
    }
}
```

#### Backward Compatibility
```csharp
// Version 1.0
public class Calculator
{
    public int Add(int a, int b) => a + b;
}

// Version 1.1 - Backward compatible
public class Calculator
{
    public int Add(int a, int b) => a + b;

    // New overload
    public double Add(double a, double b) => a + b;

    // New method
    public int Multiply(int a, int b) => a * b;
}

// Version 2.0 - Breaking changes
public class Calculator
{
    [Obsolete("Use AddAsync instead. This method will be removed in version 3.0.")]
    public int Add(int a, int b) => AddAsync(a, b).Result;

    // New async version
    public Task<int> AddAsync(int a, int b) => Task.FromResult(a + b);
}
```

### Performance Optimization

#### Package Size Optimization
```xml
<PropertyGroup>
  <!-- Minimize package size -->
  <DebugType>portable</DebugType>
  <DebugSymbols>true</DebugSymbols>

  <!-- Exclude unnecessary files -->
  <DefaultItemExcludes>$(DefaultItemExcludes);*.pdb;*.xml</DefaultItemExcludes>

  <!-- Optimize assemblies -->
  <Optimize>true</Optimize>
  <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
</PropertyGroup>

<!-- Exclude development dependencies from package -->
<ItemGroup>
  <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.5.0">
    <PrivateAssets>all</PrivateAssets>
  </PackageReference>
</ItemGroup>
```

#### Runtime Performance
```csharp
// Good: Minimize allocations
public static class StringExtensions
{
    private static readonly ConcurrentDictionary<string, string> _cache = new();

    public static string ToTitleCase(this string input)
    {
        if (string.IsNullOrEmpty(input))
            return input;

        return _cache.GetOrAdd(input, key =>
            CultureInfo.CurrentCulture.TextInfo.ToTitleCase(key.ToLower()));
    }
}

// Good: Use span and memory for better performance
public static class DataProcessor
{
    public static int ProcessData(ReadOnlySpan<byte> data)
    {
        var result = 0;
        for (var i = 0; i < data.Length; i++)
        {
            result += data[i];
        }
        return result;
    }
}
```

### Testing Strategy

#### Package Testing
```csharp
// Integration tests for package functionality
public class PackageIntegrationTests
{
    [Fact]
    public async Task Package_Should_Load_Successfully()
    {
        // Arrange
        var services = new ServiceCollection();
        services.AddMyLibrary(options =>
        {
            options.ConnectionString = "test-connection";
        });

        var provider = services.BuildServiceProvider();

        // Act
        var service = provider.GetService<IMyService>();

        // Assert
        service.Should().NotBeNull();
        var result = await service.PerformOperationAsync();
        result.Should().BeTrue();
    }

    [Theory]
    [InlineData("1.0.0")]
    [InlineData("1.1.0")]
    [InlineData("2.0.0")]
    public void Package_Should_Support_Version(string version)
    {
        // Test version compatibility
        var assembly = Assembly.LoadFrom($"MyLibrary.{version}.dll");
        var service = Activator.CreateInstance(assembly.GetType("MyLibrary.MyService"));

        service.Should().NotBeNull();
    }
}
```

## Private Repositories

### Azure Artifacts

#### Setup & Configuration
```xml
<!-- nuget.config for Azure Artifacts -->
<configuration>
  <packageSources>
    <clear />
    <add key="AzureArtifacts" value="https://pkgs.dev.azure.com/myorg/_packaging/mycompany/nuget/v3/index.json" />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
  </packageSources>

  <packageSourceCredentials>
    <AzureArtifacts>
      <add key="Username" value="any" />
      <add key="ClearTextPassword" value="%AZURE_ARTIFACTS_PAT%" />
    </AzureArtifacts>
  </packageSourceCredentials>

  <packageSourceMapping>
    <packageSource key="AzureArtifacts">
      <package pattern="MyCompany.*" />
      <package pattern="Internal.*" />
    </packageSource>
    <packageSource key="nuget.org">
      <package pattern="*" />
    </packageSource>
  </packageSourceMapping>
</configuration>
```

#### Publishing to Azure Artifacts
```bash
# Install Azure Artifacts credential provider
iex "& { $(irm https://aka.ms/install-artifacts-credprovider.ps1) } -AddNetfx"

# Set PAT environment variable
$env:AZURE_ARTIFACTS_PAT = "your-personal-access-token"

# Build and push
dotnet pack -c Release
dotnet nuget push .\bin\Release\MyPackage.1.0.0.nupkg --source "AzureArtifacts"
```

### GitHub Packages

#### GitHub Packages Configuration
```xml
<!-- nuget.config for GitHub Packages -->
<configuration>
  <packageSources>
    <add key="github" value="https://nuget.pkg.github.com/myorg/index.json" />
  </packageSources>

  <packageSourceCredentials>
    <github>
      <add key="Username" value="myusername" />
      <add key="ClearTextPassword" value="%GITHUB_TOKEN%" />
    </github>
  </packageSourceCredentials>
</configuration>
```

#### GitHub Actions Workflow
```yaml
# .github/workflows/publish.yml
name: Publish Package

on:
  release:
    types: [created]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Setup .NET
      uses: actions/setup-dotnet@v3
      with:
        dotnet-version: '6.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Test
      run: dotnet test --configuration Release --no-build --verbosity normal

    - name: Pack
      run: dotnet pack --configuration Release --no-build --output ./artifacts

    - name: Publish to GitHub Packages
      run: dotnet nuget push ./artifacts/*.nupkg --source https://nuget.pkg.github.com/myorg/index.json --api-key ${{ secrets.GITHUB_TOKEN }}

    - name: Publish to NuGet
      run: dotnet nuget push ./artifacts/*.nupkg --source https://api.nuget.org/v3/index.json --api-key ${{ secrets.NUGET_API_KEY }}
```

### MyGet & ProGet

#### MyGet Configuration
```bash
# Add MyGet feed
nuget sources add -name "MyGet" -source "https://www.myget.org/F/mycompany/api/v3/index.json"

# Push with API key
nuget push MyPackage.1.0.0.nupkg -Source MyGet -ApiKey "your-api-key"
```

#### ProGet Configuration
```xml
<configuration>
  <packageSources>
    <add key="ProGet" value="http://proget.company.com/nuget/internal/" />
  </packageSources>

  <packageSourceCredentials>
    <ProGet>
      <add key="Username" value="api" />
      <add key="ClearTextPassword" value="your-api-key" />
    </ProGet>
  </packageSourceCredentials>
</configuration>
```

## Analysis & Tooling

### Package Analysis Tools

#### NuGet Package Explorer
```bash
# Install NuGet Package Explorer
winget install Microsoft.NuGetPackageExplorer

# CLI usage
npe MyPackage.1.0.0.nupkg
```

#### Package Validation
```bash
# Install package validation tool
dotnet tool install -g Microsoft.DotNet.PackageValidation

# Validate package
dotnet package validate MyPackage.1.0.0.nupkg

# Validate against baseline
dotnet package validate MyPackage.1.0.0.nupkg --baseline-package MyPackage.0.9.0.nupkg
```

#### NuGet Trends & Insights
```bash
# Install NuGet insights tool
dotnet tool install -g NuGet.Insights

# Analyze package usage
nuget-insights analyze MyPackage --version 1.0.0

# Get download statistics
nuget-insights downloads MyPackage --range 30d
```

### Automated Quality Checks

#### GitHub Actions Quality Pipeline
```yaml
name: Package Quality Check

on:
  pull_request:
    paths: ['src/**', '*.csproj', 'Directory.Packages.props']

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0

    - name: Setup .NET
      uses: actions/setup-dotnet@v3
      with:
        dotnet-version: '6.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Test
      run: dotnet test --configuration Release --no-build --collect:"XPlat Code Coverage"

    - name: Pack
      run: dotnet pack --configuration Release --no-build --output ./artifacts

    - name: Package Validation
      run: |
        dotnet tool install -g Microsoft.DotNet.PackageValidation
        dotnet package validate ./artifacts/*.nupkg

    - name: Security Scan
      run: |
        dotnet list package --vulnerable --include-transitive
        dotnet list package --deprecated

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./TestResults/*/coverage.cobertura.xml
```

#### MSBuild Validation Targets
```xml
<!-- Directory.Build.targets -->
<Project>
  <!-- Package validation on build -->
  <Target Name="ValidatePackageOutput" AfterTargets="Pack" Condition="'$(Configuration)' == 'Release'">
    <ItemGroup>
      <PackageFiles Include="$(OutputPath)*.nupkg" />
    </ItemGroup>

    <Exec Command="dotnet package validate %(PackageFiles.Identity)"
          ContinueOnError="false" />
  </Target>

  <!-- Security checks -->
  <Target Name="SecurityCheck" BeforeTargets="Build">
    <Exec Command="dotnet list package --vulnerable --include-transitive"
          ContinueOnError="false" />
  </Target>
</Project>
```

### Monitoring & Analytics

#### Package Health Monitoring
```csharp
// Package health monitoring service
public class PackageHealthMonitor
{
    private readonly HttpClient _httpClient;
    private readonly ILogger<PackageHealthMonitor> _logger;

    public async Task<PackageHealthReport> CheckHealthAsync(string packageId)
    {
        var report = new PackageHealthReport { PackageId = packageId };

        // Check download stats
        report.Downloads = await GetDownloadStatsAsync(packageId);

        // Check vulnerabilities
        report.Vulnerabilities = await CheckVulnerabilitiesAsync(packageId);

        // Check dependencies
        report.DependencyHealth = await CheckDependencyHealthAsync(packageId);

        // Check license compliance
        report.LicenseCompliance = await CheckLicenseComplianceAsync(packageId);

        return report;
    }

    private async Task<DownloadStats> GetDownloadStatsAsync(string packageId)
    {
        var response = await _httpClient.GetAsync(
            $"https://api.nuget.org/v3-flatcontainer/{packageId.ToLower()}/index.json");

        if (response.IsSuccessStatusCode)
        {
            var content = await response.Content.ReadAsStringAsync();
            var data = JsonSerializer.Deserialize<NuGetIndexResponse>(content);
            return new DownloadStats { TotalDownloads = data.TotalDownloads };
        }

        return new DownloadStats();
    }
}
```

## Troubleshooting

### Common Issues

#### Package Resolution Issues
```bash
# Clear all caches
dotnet nuget locals all --clear

# Verbose restore
dotnet restore --verbosity detailed

# Force package restore
dotnet restore --force --no-cache

# Check package sources
dotnet nuget list source
```

#### Version Conflicts
```xml
<!-- Force specific version -->
<PackageReference Include="Newtonsoft.Json" Version="13.0.3" />

<!-- Exclude transitive dependency -->
<PackageReference Include="SomePackage" Version="1.0.0">
  <ExcludeAssets>runtime</ExcludeAssets>
</PackageReference>

<!-- Override transitive dependency -->
<PackageReference Include="TransitiveDependency" Version="2.0.0" />
```

#### Build Issues
```xml
<!-- Debugging MSBuild -->
<PropertyGroup>
  <MSBuildVerbosity>diagnostic</MSBuildVerbosity>
  <PackageValidationVerbosity>detailed</PackageValidationVerbosity>
</PropertyGroup>

<!-- Common fixes -->
<PropertyGroup>
  <RestorePackagesWithLockFile>true</RestorePackagesWithLockFile>
  <RestoreLockedMode Condition="$(CI) == 'true'">true</RestoreLockedMode>
  <DisableImplicitNuGetFallbackFolder>true</DisableImplicitNuGetFallbackFolder>
</PropertyGroup>
```

### Diagnostic Commands
```bash
# Package information
dotnet nuget why MyProject.csproj Microsoft.Extensions.Logging

# Dependency tree
dotnet list package --include-transitive --format json

# Package sources
dotnet nuget list source

# Configuration
dotnet nuget config -ShowAll

# Restore dry run
dotnet restore --verbosity detailed --no-op
```

This comprehensive guide covers all aspects of C# package management and NuGet, providing practical examples and real-world scenarios for professional development workflows.