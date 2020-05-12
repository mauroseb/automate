// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/external/infra_proxy/request/orgs.proto

package request

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type CreateOrg struct {
	// Chef organization ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Chef organization name.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Chef organization admin user.
	AdminUser string `protobuf:"bytes,3,opt,name=admin_user,json=adminUser,proto3" json:"admin_user,omitempty"`
	// Chef organization admin key.
	AdminKey string `protobuf:"bytes,4,opt,name=admin_key,json=adminKey,proto3" json:"admin_key,omitempty"`
	// Chef Infra Server ID.
	ServerId string `protobuf:"bytes,5,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	// List of projects this chef organization belongs to. May be empty.
	Projects             []string `protobuf:"bytes,6,rep,name=projects,proto3" json:"projects,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateOrg) Reset()         { *m = CreateOrg{} }
func (m *CreateOrg) String() string { return proto.CompactTextString(m) }
func (*CreateOrg) ProtoMessage()    {}
func (*CreateOrg) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e85f36e7807bf18, []int{0}
}

func (m *CreateOrg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateOrg.Unmarshal(m, b)
}
func (m *CreateOrg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateOrg.Marshal(b, m, deterministic)
}
func (m *CreateOrg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateOrg.Merge(m, src)
}
func (m *CreateOrg) XXX_Size() int {
	return xxx_messageInfo_CreateOrg.Size(m)
}
func (m *CreateOrg) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateOrg.DiscardUnknown(m)
}

var xxx_messageInfo_CreateOrg proto.InternalMessageInfo

func (m *CreateOrg) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CreateOrg) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *CreateOrg) GetAdminUser() string {
	if m != nil {
		return m.AdminUser
	}
	return ""
}

func (m *CreateOrg) GetAdminKey() string {
	if m != nil {
		return m.AdminKey
	}
	return ""
}

func (m *CreateOrg) GetServerId() string {
	if m != nil {
		return m.ServerId
	}
	return ""
}

func (m *CreateOrg) GetProjects() []string {
	if m != nil {
		return m.Projects
	}
	return nil
}

type UpdateOrg struct {
	// Chef organization ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Chef organization name.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Chef organization admin user.
	AdminUser string `protobuf:"bytes,3,opt,name=admin_user,json=adminUser,proto3" json:"admin_user,omitempty"`
	// Chef organization admin key.
	AdminKey string `protobuf:"bytes,4,opt,name=admin_key,json=adminKey,proto3" json:"admin_key,omitempty"`
	// Chef Infra Server ID.
	ServerId string `protobuf:"bytes,5,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	// List of projects this chef organization belongs to. May be empty.
	Projects             []string `protobuf:"bytes,6,rep,name=projects,proto3" json:"projects,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateOrg) Reset()         { *m = UpdateOrg{} }
func (m *UpdateOrg) String() string { return proto.CompactTextString(m) }
func (*UpdateOrg) ProtoMessage()    {}
func (*UpdateOrg) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e85f36e7807bf18, []int{1}
}

func (m *UpdateOrg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateOrg.Unmarshal(m, b)
}
func (m *UpdateOrg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateOrg.Marshal(b, m, deterministic)
}
func (m *UpdateOrg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateOrg.Merge(m, src)
}
func (m *UpdateOrg) XXX_Size() int {
	return xxx_messageInfo_UpdateOrg.Size(m)
}
func (m *UpdateOrg) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateOrg.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateOrg proto.InternalMessageInfo

func (m *UpdateOrg) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *UpdateOrg) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UpdateOrg) GetAdminUser() string {
	if m != nil {
		return m.AdminUser
	}
	return ""
}

func (m *UpdateOrg) GetAdminKey() string {
	if m != nil {
		return m.AdminKey
	}
	return ""
}

func (m *UpdateOrg) GetServerId() string {
	if m != nil {
		return m.ServerId
	}
	return ""
}

func (m *UpdateOrg) GetProjects() []string {
	if m != nil {
		return m.Projects
	}
	return nil
}

type DeleteOrg struct {
	// Chef organization ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Chef Infra Server ID.
	ServerId             string   `protobuf:"bytes,2,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteOrg) Reset()         { *m = DeleteOrg{} }
func (m *DeleteOrg) String() string { return proto.CompactTextString(m) }
func (*DeleteOrg) ProtoMessage()    {}
func (*DeleteOrg) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e85f36e7807bf18, []int{2}
}

func (m *DeleteOrg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteOrg.Unmarshal(m, b)
}
func (m *DeleteOrg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteOrg.Marshal(b, m, deterministic)
}
func (m *DeleteOrg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteOrg.Merge(m, src)
}
func (m *DeleteOrg) XXX_Size() int {
	return xxx_messageInfo_DeleteOrg.Size(m)
}
func (m *DeleteOrg) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteOrg.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteOrg proto.InternalMessageInfo

func (m *DeleteOrg) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DeleteOrg) GetServerId() string {
	if m != nil {
		return m.ServerId
	}
	return ""
}

type GetOrgs struct {
	// Chef Infra Server ID.
	ServerId             string   `protobuf:"bytes,1,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetOrgs) Reset()         { *m = GetOrgs{} }
func (m *GetOrgs) String() string { return proto.CompactTextString(m) }
func (*GetOrgs) ProtoMessage()    {}
func (*GetOrgs) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e85f36e7807bf18, []int{3}
}

func (m *GetOrgs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetOrgs.Unmarshal(m, b)
}
func (m *GetOrgs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetOrgs.Marshal(b, m, deterministic)
}
func (m *GetOrgs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetOrgs.Merge(m, src)
}
func (m *GetOrgs) XXX_Size() int {
	return xxx_messageInfo_GetOrgs.Size(m)
}
func (m *GetOrgs) XXX_DiscardUnknown() {
	xxx_messageInfo_GetOrgs.DiscardUnknown(m)
}

var xxx_messageInfo_GetOrgs proto.InternalMessageInfo

func (m *GetOrgs) GetServerId() string {
	if m != nil {
		return m.ServerId
	}
	return ""
}

type GetOrg struct {
	// Chef organization ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Chef Infra Server ID.
	ServerId             string   `protobuf:"bytes,2,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetOrg) Reset()         { *m = GetOrg{} }
func (m *GetOrg) String() string { return proto.CompactTextString(m) }
func (*GetOrg) ProtoMessage()    {}
func (*GetOrg) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e85f36e7807bf18, []int{4}
}

func (m *GetOrg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetOrg.Unmarshal(m, b)
}
func (m *GetOrg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetOrg.Marshal(b, m, deterministic)
}
func (m *GetOrg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetOrg.Merge(m, src)
}
func (m *GetOrg) XXX_Size() int {
	return xxx_messageInfo_GetOrg.Size(m)
}
func (m *GetOrg) XXX_DiscardUnknown() {
	xxx_messageInfo_GetOrg.DiscardUnknown(m)
}

var xxx_messageInfo_GetOrg proto.InternalMessageInfo

func (m *GetOrg) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *GetOrg) GetServerId() string {
	if m != nil {
		return m.ServerId
	}
	return ""
}

func init() {
	proto.RegisterType((*CreateOrg)(nil), "chef.automate.api.infra_proxy.request.CreateOrg")
	proto.RegisterType((*UpdateOrg)(nil), "chef.automate.api.infra_proxy.request.UpdateOrg")
	proto.RegisterType((*DeleteOrg)(nil), "chef.automate.api.infra_proxy.request.DeleteOrg")
	proto.RegisterType((*GetOrgs)(nil), "chef.automate.api.infra_proxy.request.GetOrgs")
	proto.RegisterType((*GetOrg)(nil), "chef.automate.api.infra_proxy.request.GetOrg")
}

func init() {
	proto.RegisterFile("api/external/infra_proxy/request/orgs.proto", fileDescriptor_6e85f36e7807bf18)
}

var fileDescriptor_6e85f36e7807bf18 = []byte{
	// 285 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x92, 0x41, 0x4b, 0x03, 0x31,
	0x14, 0x84, 0xd9, 0x6d, 0xad, 0xdd, 0x77, 0xf0, 0x90, 0x53, 0x50, 0x84, 0x52, 0x50, 0x04, 0x21,
	0x39, 0x88, 0xa0, 0x78, 0x53, 0x41, 0xc4, 0x43, 0x41, 0xe8, 0xc5, 0xcb, 0x92, 0x36, 0xaf, 0xdb,
	0x68, 0x77, 0x13, 0x5f, 0xb2, 0xd2, 0xfe, 0x1f, 0x7f, 0xa8, 0x6c, 0xd6, 0x82, 0x2b, 0x82, 0x78,
	0xf3, 0x96, 0xcc, 0x37, 0x6f, 0x98, 0xc3, 0xc0, 0xa9, 0x72, 0x46, 0xe2, 0x3a, 0x20, 0x55, 0x6a,
	0x25, 0x4d, 0xb5, 0x20, 0x95, 0x3b, 0xb2, 0xeb, 0x8d, 0x24, 0x7c, 0xad, 0xd1, 0x07, 0x69, 0xa9,
	0xf0, 0xc2, 0x91, 0x0d, 0x96, 0x1d, 0xcd, 0x97, 0xb8, 0x10, 0xaa, 0x0e, 0xb6, 0x54, 0x01, 0x85,
	0x72, 0x46, 0x7c, 0xb9, 0x10, 0x9f, 0x17, 0xe3, 0xf7, 0x04, 0xb2, 0x1b, 0x42, 0x15, 0x70, 0x42,
	0x05, 0xdb, 0x83, 0xd4, 0x68, 0x9e, 0x8c, 0x92, 0x93, 0xec, 0x31, 0x35, 0x9a, 0x31, 0xe8, 0x57,
	0xaa, 0x44, 0x9e, 0x46, 0x25, 0xbe, 0xd9, 0x21, 0x80, 0xd2, 0xa5, 0xa9, 0xf2, 0xda, 0x23, 0xf1,
	0x5e, 0x24, 0x59, 0x54, 0xa6, 0x1e, 0x89, 0x1d, 0x40, 0xfb, 0xc9, 0x5f, 0x70, 0xc3, 0xfb, 0x91,
	0x0e, 0xa3, 0xf0, 0x80, 0x9b, 0x06, 0x7a, 0xa4, 0x37, 0xa4, 0xdc, 0x68, 0xbe, 0xd3, 0xc2, 0x56,
	0xb8, 0xd7, 0x6c, 0x1f, 0x86, 0x8e, 0xec, 0x33, 0xce, 0x83, 0xe7, 0x83, 0x51, 0xaf, 0x61, 0xdb,
	0x7f, 0xac, 0x39, 0x75, 0xfa, 0xbf, 0xd7, 0xbc, 0x80, 0xec, 0x16, 0x57, 0xf8, 0x73, 0xcb, 0x4e,
	0x6a, 0xda, 0x4d, 0x1d, 0x1f, 0xc3, 0xee, 0x1d, 0x86, 0x09, 0x15, 0xbe, 0xeb, 0x4b, 0xbe, 0xf9,
	0xce, 0x61, 0xd0, 0xfa, 0xfe, 0x14, 0x7f, 0x7d, 0xf5, 0x74, 0x59, 0x98, 0xb0, 0xac, 0x67, 0x62,
	0x6e, 0x4b, 0xd9, 0x4c, 0x43, 0x6e, 0xa7, 0x21, 0x7f, 0x5b, 0xd5, 0x6c, 0x10, 0x17, 0x75, 0xf6,
	0x11, 0x00, 0x00, 0xff, 0xff, 0x04, 0x56, 0x71, 0x28, 0x80, 0x02, 0x00, 0x00,
}
