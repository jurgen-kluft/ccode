package axe

type XcodeProductTypeEnum string

const (
	XcodeProductTypeApplication XcodeProductTypeEnum = "com.apple.product-type.application"
	XcodeProductTypeFramework   XcodeProductTypeEnum = "com.apple.product-type.framework"
	XcodeProductTypeStaticLib   XcodeProductTypeEnum = "com.apple.product-type.library.static"
	XcodeProductTypeDynamicLib  XcodeProductTypeEnum = "com.apple.product-type.library.dynamic"
	XcodeProductTypeTool        XcodeProductTypeEnum = "com.apple.product-type.tool"
)

func (p XcodeProductTypeEnum) String() string {
	return string(p)
}

const (
	XcodeBuildPhaseSourcesUUID    = "B0000000B0000000B0000002"
	XcodeBuildPhaseFrameworksUUID = "B0000000B0000000B0000003"
	XcodeBuildPhaseHeadersUUID    = "B0000000B0000000B0000004"
	XcodeBuildPhaseResourcesUUID  = "B0000000B0000000B0000005"
	XcodeMainGroupUUID            = "C0000000B0000000B0000001"
	XcodeProductGroupUUID         = "C0000000B0000000B0000002"
	XcodeDependenciesGroupUUID    = "C0000000B0000000B0000003"
	XcodeResourcesGroupUUID       = "C0000000B0000000B0000004"
	XcodeKSourceTreeProject       = "SOURCE_ROOT"
	XcodeKSourceTreeGroup         = "\"<group>\""
	XcodeKSourceTreeAbsolute      = "\"<absolute>\""
)
