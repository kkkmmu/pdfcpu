/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pdfcpu

import (
	"github.com/hhrutter/pdfcpu/pkg/filter"
	"github.com/pkg/errors"
)

const (

	// ExcludePatternCS ...
	ExcludePatternCS = true

	// IncludePatternCS ...
	IncludePatternCS = false

	isAlternateImageStreamDict   = true
	isNoAlternateImageStreamDict = false
)

func validateReferenceDictPageEntry(xRefTable *XRefTable, obj PDFObject) error {

	obj, err := xRefTable.Dereference(obj)
	if err != nil || obj == nil {
		return err
	}

	switch obj.(type) {

	case PDFInteger, PDFStringLiteral, PDFHexLiteral:
		// no further processing

	default:
		return errors.New("validateReferenceDictPageEntry: corrupt type")

	}

	return nil
}

func validateReferenceDict(xRefTable *XRefTable, dict *PDFDict) error {

	// see 8.10.4 Reference XObjects

	dictName := "refDict"

	// F, file spec, required
	_, err := validateFileSpecEntry(xRefTable, dict, dictName, "F", REQUIRED, V10)
	if err != nil {
		return err
	}

	// Page, integer or text string, required
	obj, ok := dict.Find("Page")
	if !ok {
		return errors.New("validateReferenceDict: missing required entry \"Page\"")
	}

	err = validateReferenceDictPageEntry(xRefTable, obj)
	if err != nil {
		return err
	}

	// ID, string array, optional
	_, err = validateStringArrayEntry(xRefTable, dict, dictName, "ID", OPTIONAL, V10, func(arr PDFArray) bool { return len(arr) == 2 })

	return err
}

func validateOPIDictV13Part1(xRefTable *XRefTable, dict *PDFDict, dictName string) error {

	// Type, optional, name
	_, err := validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, V10, func(s string) bool { return s == "OPI" })
	if err != nil {
		return err
	}

	// Version, required, number
	_, err = validateNumberEntry(xRefTable, dict, dictName, "Version", REQUIRED, V10, func(f float64) bool { return f == 1.3 })
	if err != nil {
		return err
	}

	// F, required, file specification
	_, err = validateFileSpecEntry(xRefTable, dict, dictName, "F", REQUIRED, V10)
	if err != nil {
		return err
	}

	// ID, optional, byte string
	_, err = validateStringEntry(xRefTable, dict, dictName, "ID", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Comments, optional, text string
	_, err = validateStringEntry(xRefTable, dict, dictName, "Comments", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Size, required, array of integers, len 2
	_, err = validateIntegerArrayEntry(xRefTable, dict, dictName, "Size", REQUIRED, V10, func(a PDFArray) bool { return len(a) == 2 })
	if err != nil {
		return err
	}

	// CropRect, required, array of integers, len 4
	_, err = validateRectangleEntry(xRefTable, dict, dictName, "CropRect", REQUIRED, V10, nil)

	if err != nil {
		return err
	}

	// CropFixed, optional, array of numbers, len 4
	_, err = validateRectangleEntry(xRefTable, dict, dictName, "CropFixed", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Position, required, array of numbers, len 8
	_, err = validateNumberArrayEntry(xRefTable, dict, dictName, "Position", REQUIRED, V10, func(a PDFArray) bool { return len(a) == 8 })

	return err
}

func validateOPIDictV13Part2(xRefTable *XRefTable, dict *PDFDict, dictName string) error {

	// Resolution, optional, array of numbers, len 2
	_, err := validateNumberArrayEntry(xRefTable, dict, dictName, "Resolution", OPTIONAL, V10, func(a PDFArray) bool { return len(a) == 2 })
	if err != nil {
		return err
	}

	// ColorType, optional, name
	_, err = validateNameEntry(xRefTable, dict, dictName, "ColorType", OPTIONAL, V10, func(s string) bool { return s == "Process" || s == "Spot" || s == "Separation" })
	if err != nil {
		return err
	}

	// Color, optional, array, len 5
	_, err = validateArrayEntry(xRefTable, dict, dictName, "Color", OPTIONAL, V10, func(a PDFArray) bool { return len(a) == 5 })
	if err != nil {
		return err
	}

	// Tint, optional, number
	_, err = validateNumberEntry(xRefTable, dict, dictName, "Tint", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Overprint, optional, boolean
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "Overprint", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// ImageType, optional, array of integers, len 2
	_, err = validateIntegerArrayEntry(xRefTable, dict, dictName, "ImageType", OPTIONAL, V10, func(a PDFArray) bool { return len(a) == 2 })
	if err != nil {
		return err
	}

	// GrayMap, optional, array of integers
	_, err = validateIntegerArrayEntry(xRefTable, dict, dictName, "GrayMap", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Transparency, optional, boolean
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "Transparency", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Tags, optional, array
	_, err = validateArrayEntry(xRefTable, dict, dictName, "Tags", OPTIONAL, V10, nil)

	return err
}

func validateOPIDictV13(xRefTable *XRefTable, dict *PDFDict) error {

	// 14.11.7 Open Prepresse interface (OPI)

	dictName := "opiDictV13"

	err := validateOPIDictV13Part1(xRefTable, dict, dictName)
	if err != nil {
		return err
	}

	return validateOPIDictV13Part2(xRefTable, dict, dictName)
}

func validateOPIDictInks(xRefTable *XRefTable, obj PDFObject) error {

	obj, err := xRefTable.Dereference(obj)
	if err != nil || obj == nil {
		return err
	}

	switch obj := obj.(type) {

	case PDFName:
		if colorant := obj.String(); colorant != "full_color" && colorant != "registration" {
			return errors.New("validateOPIDictInks: corrupt colorant name")
		}

	case PDFArray:
		// no further processing

	default:
		return errors.New("validateOPIDictInks: corrupt type")

	}

	return nil
}

func validateOPIDictV20(xRefTable *XRefTable, dict *PDFDict) error {

	// 14.11.7 Open Prepresse interface (OPI)

	dictName := "opiDictV20"

	_, err := validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, V10, func(s string) bool { return s == "OPI" })
	if err != nil {
		return err
	}

	_, err = validateNumberEntry(xRefTable, dict, dictName, "Version", REQUIRED, V10, func(f float64) bool { return f == 2.0 })
	if err != nil {
		return err
	}

	_, err = validateFileSpecEntry(xRefTable, dict, dictName, "F", REQUIRED, V10)
	if err != nil {
		return err
	}

	_, err = validateStringEntry(xRefTable, dict, dictName, "MainImage", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	_, err = validateArrayEntry(xRefTable, dict, dictName, "Tags", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	_, err = validateNumberArrayEntry(xRefTable, dict, dictName, "Size", OPTIONAL, V10, func(arr PDFArray) bool { return len(arr) == 2 })
	if err != nil {
		return err
	}

	_, err = validateRectangleEntry(xRefTable, dict, dictName, "CropRect", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	_, err = validateBooleanEntry(xRefTable, dict, dictName, "Overprint", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	if obj, found := dict.Find("Inks"); found {
		err = validateOPIDictInks(xRefTable, obj)
		if err != nil {
			return err
		}
	}

	_, err = validateIntegerArrayEntry(xRefTable, dict, dictName, "IncludedImageDimensions", OPTIONAL, V10, func(arr PDFArray) bool { return len(arr) == 2 })
	if err != nil {
		return err
	}

	_, err = validateIntegerEntry(xRefTable, dict, dictName, "IncludedImageQuality", OPTIONAL, V10, func(i int) bool { return i >= 1 && i <= 3 })

	return err
}

func validateOPIVersionDict(xRefTable *XRefTable, dict *PDFDict) error {

	// 14.11.7 Open Prepresse interface (OPI)

	if dict.Len() != 1 {
		return errors.New("validateOPIVersionDict: must have exactly one entry keyed 1.3 or 2.0")
	}

	validateOPIVersion := func(s string) bool { return memberOf(s, []string{"1.3", "2.0"}) }

	for opiVersion, obj := range dict.Dict {

		if !validateOPIVersion(opiVersion) {
			return errors.New("validateOPIVersionDict: invalid OPI version")
		}

		dict, err := xRefTable.DereferenceDict(obj)
		if err != nil || dict == nil {
			return err
		}

		if opiVersion == "1.3" {
			err = validateOPIDictV13(xRefTable, dict)
		} else {
			err = validateOPIDictV20(xRefTable, dict)
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func validateMaskStreamDict(xRefTable *XRefTable, streamDict *PDFStreamDict) error {

	if streamDict.Type() != nil && *streamDict.Type() != "XObject" {
		return errors.New("validateMaskStreamDict: corrupt imageStreamDict type")
	}

	if streamDict.Subtype() == nil || *streamDict.Subtype() != "Image" {
		return errors.New("validateMaskStreamDict: corrupt imageStreamDict subtype")
	}

	return validateImageStreamDict(xRefTable, streamDict, isNoAlternateImageStreamDict)
}

func validateMaskEntry(xRefTable *XRefTable, dict *PDFDict, dictName, entryName string, required bool, sinceVersion PDFVersion) error {

	// stream ("explicit masking", another Image XObject) or array of colors ("color key masking")

	obj, err := validateEntry(xRefTable, dict, dictName, entryName, required, sinceVersion)
	if err != nil || obj == nil {
		return err
	}

	switch obj := obj.(type) {

	case PDFStreamDict:
		err = validateMaskStreamDict(xRefTable, &obj)
		if err != nil {
			return err
		}

	case PDFArray:
		// no further processing

	default:

		return errors.Errorf("validateMaskEntry: dict=%s corrupt entry \"%s\"\n", dictName, entryName)

	}

	return nil
}

func validateAlternateImageStreamDicts(xRefTable *XRefTable, dict *PDFDict, dictName string, entryName string, required bool, sinceVersion PDFVersion) error {

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, nil)
	if err != nil {
		return err
	}
	if arr == nil {
		if required {
			return errors.Errorf("validateAlternateImageStreamDicts: dict=%s required entry \"%s\" missing.", dictName, entryName)
		}
		return nil
	}

	for _, obj := range *arr {

		streamDict, err := validateStreamDict(xRefTable, obj)
		if err != nil {
			return err
		}

		if streamDict == nil {
			continue
		}

		err = validateImageStreamDict(xRefTable, streamDict, isAlternateImageStreamDict)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateImageStreamDictPart1(xRefTable *XRefTable, streamDict *PDFStreamDict, dictName string) (isImageMask bool, err error) {

	dict := streamDict.PDFDict

	// Width, integer, required
	_, err = validateIntegerEntry(xRefTable, &dict, dictName, "Width", REQUIRED, V10, nil)
	if err != nil {
		return false, err
	}

	// Height, integer, required
	_, err = validateIntegerEntry(xRefTable, &dict, dictName, "Height", REQUIRED, V10, nil)
	if err != nil {
		return false, err
	}

	// ImageMask, boolean, optional
	imageMask, err := validateBooleanEntry(xRefTable, &dict, dictName, "ImageMask", OPTIONAL, V10, nil)
	if err != nil {
		return false, err
	}

	isImageMask = imageMask != nil && *imageMask == true

	// ColorSpace, name or array, required unless used filter is JPXDecode; not allowed for imagemasks.
	if !isImageMask {

		required := REQUIRED

		if streamDict.HasSoleFilterNamed(filter.JPX) {
			required = OPTIONAL
		}

		if streamDict.HasSoleFilterNamed(filter.CCITTFax) && xRefTable.ValidationMode == ValidationRelaxed {
			required = OPTIONAL
		}

		err = validateColorSpaceEntry(xRefTable, &dict, dictName, "ColorSpace", required, ExcludePatternCS)
		if err != nil {
			return false, err
		}

	}

	return isImageMask, nil
}

func validateImageStreamDictPart2(xRefTable *XRefTable, streamDict *PDFStreamDict, dictName string, isImageMask, isAlternate bool) error {

	dict := streamDict.PDFDict

	// BitsPerComponent, integer
	required := REQUIRED
	if streamDict.HasSoleFilterNamed(filter.JPX) || isImageMask {
		required = OPTIONAL
	}
	// For imageMasks BitsPerComponent must be 1.
	var validateBPC func(i int) bool
	if isImageMask {
		validateBPC = func(i int) bool {
			return i == 1
		}
	}
	_, err := validateIntegerEntry(xRefTable, &dict, dictName, "BitsPerComponent", required, V10, validateBPC)
	if err != nil {
		return err
	}

	// Intent, name, optional, since V1.0
	validate := func(s string) bool {
		return memberOf(s, []string{"AbsoluteColorimetric", "RelativeColorimetric", "Saturation", "Perceptual"})
	}
	_, err = validateNameEntry(xRefTable, &dict, dictName, "Intent", OPTIONAL, V11, validate)
	if err != nil {
		return err
	}

	// Mask, stream or array, optional since V1.3; not allowed for image masks.
	if !isImageMask {
		err = validateMaskEntry(xRefTable, &dict, dictName, "Mask", OPTIONAL, V13)
		if err != nil {
			return err
		}
	}

	// Decode, array, optional
	_, err = validateNumberArrayEntry(xRefTable, &dict, dictName, "Decode", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Interpolate, boolean, optional
	_, err = validateBooleanEntry(xRefTable, &dict, dictName, "Interpolate", OPTIONAL, V10, nil)
	if err != nil {
		return err
	}

	// Alternates, array, optional, since V1.3
	if !isAlternate {
		err = validateAlternateImageStreamDicts(xRefTable, &dict, dictName, "Alternates", OPTIONAL, V13)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateImageStreamDict(xRefTable *XRefTable, streamDict *PDFStreamDict, isAlternate bool) error {

	dict := streamDict.PDFDict

	dictName := "imageStreamDict"

	var isImageMask bool

	isImageMask, err := validateImageStreamDictPart1(xRefTable, streamDict, dictName)
	if err != nil {
		return err
	}

	err = validateImageStreamDictPart2(xRefTable, streamDict, dictName, isImageMask, isAlternate)
	if err != nil {
		return err
	}

	// SMask, stream, optional, since V1.4
	sinceVersion := V14
	if xRefTable.ValidationMode == ValidationRelaxed {
		sinceVersion = V13
	}
	sd, err := validateStreamDictEntry(xRefTable, &dict, dictName, "SMask", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	if sd != nil {
		err = validateImageStreamDict(xRefTable, sd, isNoAlternateImageStreamDict)
		if err != nil {
			return err
		}
	}

	// SMaskInData, integer, optional
	_, err = validateIntegerEntry(xRefTable, &dict, dictName, "SMaskInData", OPTIONAL, V10, func(i int) bool { return i >= 0 && i <= 2 })
	if err != nil {
		return err
	}

	// Name, name, required for V10
	// Shall no longer be used.
	_, err = validateNameEntry(xRefTable, &dict, dictName, "Name", xRefTable.Version() == V10, V10, nil)
	if err != nil {
		return err
	}

	// StructParent, integer, optional
	_, err = validateIntegerEntry(xRefTable, &dict, dictName, "StructParent", OPTIONAL, V13, nil)
	if err != nil {
		return err
	}

	// ID, byte string, optional, since V1.3
	_, err = validateStringEntry(xRefTable, &dict, dictName, "ID", OPTIONAL, V13, nil)
	if err != nil {
		return err
	}

	// OPI, dict, optional since V1.2
	err = validateEntryOPI(xRefTable, &dict, dictName, "OPI", OPTIONAL, V12)
	if err != nil {
		return err
	}

	// Metadata, stream, optional since V1.4
	err = validateMetadata(xRefTable, &streamDict.PDFDict, OPTIONAL, V14)
	if err != nil {
		return err
	}

	// OC, dict, optional since V1.5
	return validateEntryOC(xRefTable, &dict, dictName, "OC", OPTIONAL, V15)
}

func validateFormStreamDictPart1(xRefTable *XRefTable, streamDict *PDFStreamDict, dictName string) error {

	dict := streamDict.PDFDict

	_, err := validateIntegerEntry(xRefTable, &dict, dictName, "FormType", OPTIONAL, V10, func(i int) bool { return i == 1 })
	if err != nil {
		return err
	}

	_, err = validateRectangleEntry(xRefTable, &dict, dictName, "BBox", REQUIRED, V10, nil)
	if err != nil {
		return err
	}

	_, err = validateNumberArrayEntry(xRefTable, &dict, dictName, "Matrix", OPTIONAL, V10, func(arr PDFArray) bool { return len(arr) == 6 })
	if err != nil {
		return err
	}

	// Resources, dict, optional, since V1.2
	if obj, ok := streamDict.Find("Resources"); ok {
		_, err = validateResourceDict(xRefTable, obj)
		if err != nil {
			return err
		}
	}

	// Group, dict, optional, since V1.4
	err = validatePageEntryGroup(xRefTable, &dict, OPTIONAL, V14)
	if err != nil {
		return err
	}

	// Ref, dict, optional, since V1.4
	d, err := validateDictEntry(xRefTable, &dict, dictName, "Ref", OPTIONAL, V14, nil)
	if err != nil {
		return err
	}
	if d != nil {
		err = validateReferenceDict(xRefTable, d)
		if err != nil {
			return err
		}
	}

	// Metadata, stream, optional, since V1.4
	return validateMetadata(xRefTable, &dict, OPTIONAL, V14)
}

func validateEntryOC(xRefTable *XRefTable, dict *PDFDict, dictName, entryName string, required bool, sinceVersion PDFVersion) error {

	var d *PDFDict

	d, err := validateDictEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, nil)
	if err != nil {
		return err
	}

	if d != nil {
		err = validateOptionalContentGroupDict(xRefTable, d, sinceVersion)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateEntryOPI(xRefTable *XRefTable, dict *PDFDict, dictName, entryName string, required bool, sinceVersion PDFVersion) error {

	d, err := validateDictEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, nil)
	if err != nil {
		return err
	}

	if d != nil {
		err = validateOPIVersionDict(xRefTable, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateFormStreamDictPart2(xRefTable *XRefTable, dict *PDFDict, dictName string) error {

	// PieceInfo, dict, optional, since V1.3
	hasPieceInfo, err := validatePieceInfo(xRefTable, dict, dictName, "PieceInfo", OPTIONAL, V13)
	if err != nil {
		return err
	}

	// LastModified, date, required if PieceInfo present, since V1.3
	lm, err := validateDateEntry(xRefTable, dict, dictName, "LastModified", OPTIONAL, V13)
	if err != nil {
		return err
	}

	if hasPieceInfo && lm == nil {
		err = errors.New("validateFormStreamDictPart2: missing \"LastModified\" (required by \"PieceInfo\")")
		return err
	}

	// StructParent, integer
	sp, err := validateIntegerEntry(xRefTable, dict, dictName, "StructParent", OPTIONAL, V13, nil)
	if err != nil {
		return err
	}

	// StructParents, integer
	sps, err := validateIntegerEntry(xRefTable, dict, dictName, "StructParents", OPTIONAL, V13, nil)
	if err != nil {
		return err
	}
	if sp != nil && sps != nil {
		return errors.New("validateFormStreamDictPart2: only \"StructParent\" or \"StructParents\" allowed")
	}

	// OPI, dict, optional, since V1.2
	err = validateEntryOPI(xRefTable, dict, dictName, "OPI", OPTIONAL, V12)
	if err != nil {
		return err
	}

	// OC, optional, content group dict or content membership dict, since V1.5
	// Specifying the optional content properties for the annotation.
	err = validateOptionalContent(xRefTable, dict, dictName, "OC", OPTIONAL, V15)
	if err != nil {
		return err
	}

	// Name, name, optional (required in 1.0)
	required := xRefTable.Version() == V10
	_, err = validateNameEntry(xRefTable, dict, dictName, "Name", required, V10, nil)
	if err != nil {
		return err
	}

	return nil
}

func validateFormStreamDict(xRefTable *XRefTable, streamDict *PDFStreamDict) error {

	// 8.10 Form XObjects

	dictName := "formStreamDict"

	err := validateFormStreamDictPart1(xRefTable, streamDict, dictName)
	if err != nil {
		return err
	}

	return validateFormStreamDictPart2(xRefTable, &streamDict.PDFDict, dictName)
}

func validateXObjectStreamDict(xRefTable *XRefTable, obj PDFObject) error {

	// see 8.8 External Objects

	sd, err := xRefTable.DereferenceStreamDict(obj)
	if err != nil || obj == nil {
		return err
	}

	d := sd.PDFDict
	dictName := "xObjectStreamDict"

	_, err = validateNameEntry(xRefTable, &d, dictName, "Type", OPTIONAL, V10, func(s string) bool { return s == "XObject" })
	if err != nil {
		return err
	}

	required := REQUIRED
	if xRefTable.ValidationMode == ValidationRelaxed {
		required = OPTIONAL
	}
	subtype, err := validateNameEntry(xRefTable, &d, dictName, "Subtype", required, V10, nil)
	if err != nil {
		return err
	}

	if subtype == nil {

		// relaxed
		_, found := sd.Find("BBox")
		if found {
			return validateFormStreamDict(xRefTable, sd)
		}

		// Relaxed for page Thumb
		return validateImageStreamDict(xRefTable, sd, isNoAlternateImageStreamDict)
	}

	switch *subtype {

	case "Form":
		err = validateFormStreamDict(xRefTable, sd)

	case "Image":
		err = validateImageStreamDict(xRefTable, sd, isNoAlternateImageStreamDict)

	case "PS":
		err = errors.Errorf("validateXObjectStreamDict: PostScript XObjects should not be used")

	default:
		return errors.Errorf("validateXObjectStreamDict: unknown Subtype: %s\n", *subtype)

	}

	return err
}

func validateGroupAttributesDict(xRefTable *XRefTable, obj PDFObject) error {

	// see 11.6.6 Transparency Group XObjects

	dict, err := xRefTable.DereferenceDict(obj)
	if err != nil || dict == nil {
		return err
	}

	dictName := "groupAttributesDict"

	// Type, name, optional
	_, err = validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, V10, func(s string) bool { return s == "Group" })
	if err != nil {
		return err
	}

	// S, name, required
	_, err = validateNameEntry(xRefTable, dict, dictName, "S", REQUIRED, V10, func(s string) bool { return s == "Transparency" })
	if err != nil {
		return err
	}

	// CS, colorSpace, optional
	err = validateColorSpaceEntry(xRefTable, dict, dictName, "CS", OPTIONAL, ExcludePatternCS)
	if err != nil {
		return err
	}

	// I, boolean, optional
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "I", OPTIONAL, V10, nil)

	return err
}

func validateXObjectResourceDict(xRefTable *XRefTable, obj PDFObject, sinceVersion PDFVersion) error {

	// Version check
	err := xRefTable.ValidateVersion("XObjectResourceDict", sinceVersion)
	if err != nil {
		return err
	}

	dict, err := xRefTable.DereferenceDict(obj)
	if err != nil || dict == nil {
		return err
	}

	// Iterate over XObject resource dictionary
	for _, obj := range dict.Dict {

		// Process XObject dict
		err = validateXObjectStreamDict(xRefTable, obj)
		if err != nil {
			return err
		}
	}

	return nil
}
