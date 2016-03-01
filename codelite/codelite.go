package codelite

import (
	"fmt"
)

func GenerateCodeLiteCppWorkspace() {
	example := `<?xml version="1.0" encoding="UTF-8"?>
<CodeLite_Workspace Name="xtest" Database="">
  <Project Name="xtest" Path="xtest.project" Active="Yes"/>
  <BuildMatrix>
    <WorkspaceConfiguration Name="Debug" Selected="yes">
      <Environment/>
      <Project Name="xtest" ConfigName="Debug"/>
    </WorkspaceConfiguration>
    <WorkspaceConfiguration Name="Release" Selected="yes">
      <Environment/>
      <Project Name="xtest" ConfigName="Release"/>
    </WorkspaceConfiguration>
  </BuildMatrix>
</CodeLite_Workspace>
`
	fmt.Println(example)
}

func GenerateCodeLiteCppProject() {
	example := `<?xml version="1.0" encoding="UTF-8"?>
<CodeLite_Project Name="xtest" InternalType="Library">
  <Description/>
  <Dependencies/>
  <Settings Type="Static Library">
    <GlobalSettings>
      <Compiler Options="" C_Options="" Assembler="">
        <IncludePath Value="."/>
      </Compiler>
      <Linker Options="">
        <LibraryPath Value="."/>
      </Linker>
      <ResourceCompiler Options=""/>
    </GlobalSettings>
    <Configuration Name="Debug" CompilerType="Visual C++ 14" DebuggerType="GNU gdb debugger" Type="Static Library" BuildCmpWithGlobalSettings="append" BuildLnkWithGlobalSettings="append" BuildResWithGlobalSettings="append">
      <Compiler Options="-g" C_Options="-g" Assembler="" Required="yes" PreCompiledHeader="" PCHInCommandLine="no" PCHFlags="" PCHFlagsPolicy="0">
        <IncludePath Value="."/>
      </Compiler>
      <Linker Options="" Required="yes"/>
      <ResourceCompiler Options="" Required="no"/>
      <General OutputFile="$(IntermediateDirectory)/lib$(ProjectName).a" IntermediateDirectory="./Debug" Command="" CommandArguments="" UseSeparateDebugArgs="no" DebugArguments="" WorkingDirectory="$(IntermediateDirectory)" PauseExecWhenProcTerminates="yes" IsGUIProgram="no" IsEnabled="yes"/>
      <Environment EnvVarSetName="&lt;Use Defaults&gt;" DbgSetName="&lt;Use Defaults&gt;">
        <![CDATA[]]>
      </Environment>
      <Debugger IsRemote="no" RemoteHostName="" RemoteHostPort="" DebuggerPath="" IsExtended="no">
        <DebuggerSearchPaths/>
        <PostConnectCommands/>
        <StartupCommands/>
      </Debugger>
      <PreBuild/>
      <PostBuild/>
      <CustomBuild Enabled="no">
        <RebuildCommand/>
        <CleanCommand/>
        <BuildCommand/>
        <PreprocessFileCommand/>
        <SingleFileCommand/>
        <MakefileGenerationCommand/>
        <ThirdPartyToolName/>
        <WorkingDirectory/>
      </CustomBuild>
      <AdditionalRules>
        <CustomPostBuild/>
        <CustomPreBuild/>
      </AdditionalRules>
      <Completion EnableCpp11="no" EnableCpp14="no">
        <ClangCmpFlagsC/>
        <ClangCmpFlags/>
        <ClangPP/>
        <SearchPaths/>
      </Completion>
    </Configuration>
    <Configuration Name="Release" CompilerType="Visual C++ 14" DebuggerType="GNU gdb debugger" Type="Static Library" BuildCmpWithGlobalSettings="append" BuildLnkWithGlobalSettings="append" BuildResWithGlobalSettings="append">
      <Compiler Options="" C_Options="" Assembler="" Required="yes" PreCompiledHeader="" PCHInCommandLine="no" PCHFlags="" PCHFlagsPolicy="0">
        <IncludePath Value="."/>
      </Compiler>
      <Linker Options="" Required="yes"/>
      <ResourceCompiler Options="" Required="no"/>
      <General OutputFile="$(IntermediateDirectory)/lib$(ProjectName).a" IntermediateDirectory="./Release" Command="" CommandArguments="" UseSeparateDebugArgs="no" DebugArguments="" WorkingDirectory="$(IntermediateDirectory)" PauseExecWhenProcTerminates="yes" IsGUIProgram="no" IsEnabled="yes"/>
      <Environment EnvVarSetName="&lt;Use Defaults&gt;" DbgSetName="&lt;Use Defaults&gt;">
        <![CDATA[]]>
      </Environment>
      <Debugger IsRemote="no" RemoteHostName="" RemoteHostPort="" DebuggerPath="" IsExtended="yes">
        <DebuggerSearchPaths/>
        <PostConnectCommands/>
        <StartupCommands/>
      </Debugger>
      <PreBuild/>
      <PostBuild/>
      <CustomBuild Enabled="no">
        <RebuildCommand/>
        <CleanCommand/>
        <BuildCommand/>
        <PreprocessFileCommand/>
        <SingleFileCommand/>
        <MakefileGenerationCommand/>
        <ThirdPartyToolName/>
        <WorkingDirectory/>
      </CustomBuild>
      <AdditionalRules>
        <CustomPostBuild/>
        <CustomPreBuild/>
      </AdditionalRules>
      <Completion EnableCpp11="no" EnableCpp14="no">
        <ClangCmpFlagsC/>
        <ClangCmpFlags/>
        <ClangPP/>
        <SearchPaths/>
      </Completion>
    </Configuration>
  </Settings>
  <VirtualDirectory Name="cpp">
    <VirtualDirectory Name="private">
      <File Name="cpp/private/private1.cpp"/>
    </VirtualDirectory>
    <File Name="cpp/file1.cpp"/>
    <File Name="cpp/file2.cpp"/>
  </VirtualDirectory>
  <VirtualDirectory Name="include">
    <VirtualDirectory Name="xtest">
      <VirtualDirectory Name="private">
        <File Name="include/xtest/private/private1.h"/>
      </VirtualDirectory>
      <File Name="include/xtest/file1.h"/>
      <File Name="include/xtest/file2.h"/>
    </VirtualDirectory>
  </VirtualDirectory>
</CodeLite_Project>
`

	fmt.Println(example)
}
