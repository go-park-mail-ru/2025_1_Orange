// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package dto

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson820b0337DecodeResuMatchInternalEntityDto(in *jlexer.Lexer, out *EmployerProfileUpdate) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "company_name":
			out.CompanyName = string(in.String())
		case "legal_address":
			out.LegalAddress = string(in.String())
		case "slogan":
			out.Slogan = string(in.String())
		case "website":
			out.Website = string(in.String())
		case "description":
			out.Description = string(in.String())
		case "vk":
			out.Vk = string(in.String())
		case "telegram":
			out.Telegram = string(in.String())
		case "facebook":
			out.Facebook = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson820b0337EncodeResuMatchInternalEntityDto(out *jwriter.Writer, in EmployerProfileUpdate) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix[1:])
		out.String(string(in.CompanyName))
	}
	{
		const prefix string = ",\"legal_address\":"
		out.RawString(prefix)
		out.String(string(in.LegalAddress))
	}
	{
		const prefix string = ",\"slogan\":"
		out.RawString(prefix)
		out.String(string(in.Slogan))
	}
	{
		const prefix string = ",\"website\":"
		out.RawString(prefix)
		out.String(string(in.Website))
	}
	{
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Description))
	}
	{
		const prefix string = ",\"vk\":"
		out.RawString(prefix)
		out.String(string(in.Vk))
	}
	{
		const prefix string = ",\"telegram\":"
		out.RawString(prefix)
		out.String(string(in.Telegram))
	}
	{
		const prefix string = ",\"facebook\":"
		out.RawString(prefix)
		out.String(string(in.Facebook))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v EmployerProfileUpdate) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson820b0337EncodeResuMatchInternalEntityDto(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v EmployerProfileUpdate) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson820b0337EncodeResuMatchInternalEntityDto(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *EmployerProfileUpdate) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson820b0337DecodeResuMatchInternalEntityDto(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *EmployerProfileUpdate) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson820b0337DecodeResuMatchInternalEntityDto(l, v)
}
func easyjson820b0337DecodeResuMatchInternalEntityDto1(in *jlexer.Lexer, out *EmployerProfileResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = int(in.Int())
		case "company_name":
			out.CompanyName = string(in.String())
		case "legal_address":
			out.LegalAddress = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "slogan":
			out.Slogan = string(in.String())
		case "website":
			out.Website = string(in.String())
		case "description":
			out.Description = string(in.String())
		case "vk":
			out.Vk = string(in.String())
		case "telegram":
			out.Telegram = string(in.String())
		case "facebook":
			out.Facebook = string(in.String())
		case "logo_path":
			out.LogoPath = string(in.String())
		case "created_at":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.CreatedAt).UnmarshalJSON(data))
			}
		case "updated_at":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.UpdatedAt).UnmarshalJSON(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson820b0337EncodeResuMatchInternalEntityDto1(out *jwriter.Writer, in EmployerProfileResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int(int(in.ID))
	}
	{
		const prefix string = ",\"company_name\":"
		out.RawString(prefix)
		out.String(string(in.CompanyName))
	}
	{
		const prefix string = ",\"legal_address\":"
		out.RawString(prefix)
		out.String(string(in.LegalAddress))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"slogan\":"
		out.RawString(prefix)
		out.String(string(in.Slogan))
	}
	{
		const prefix string = ",\"website\":"
		out.RawString(prefix)
		out.String(string(in.Website))
	}
	{
		const prefix string = ",\"description\":"
		out.RawString(prefix)
		out.String(string(in.Description))
	}
	{
		const prefix string = ",\"vk\":"
		out.RawString(prefix)
		out.String(string(in.Vk))
	}
	{
		const prefix string = ",\"telegram\":"
		out.RawString(prefix)
		out.String(string(in.Telegram))
	}
	{
		const prefix string = ",\"facebook\":"
		out.RawString(prefix)
		out.String(string(in.Facebook))
	}
	{
		const prefix string = ",\"logo_path\":"
		out.RawString(prefix)
		out.String(string(in.LogoPath))
	}
	{
		const prefix string = ",\"created_at\":"
		out.RawString(prefix)
		out.Raw((in.CreatedAt).MarshalJSON())
	}
	{
		const prefix string = ",\"updated_at\":"
		out.RawString(prefix)
		out.Raw((in.UpdatedAt).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v EmployerProfileResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson820b0337EncodeResuMatchInternalEntityDto1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v EmployerProfileResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson820b0337EncodeResuMatchInternalEntityDto1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *EmployerProfileResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson820b0337DecodeResuMatchInternalEntityDto1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *EmployerProfileResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson820b0337DecodeResuMatchInternalEntityDto1(l, v)
}
