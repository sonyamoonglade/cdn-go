package modules

import (
	"bytes"

	module_errors "animakuro/cdn/internal/modules/errors"

	"github.com/h2non/bimg"
)

const (
	yAxisResizePercent = 0.2114
)

const (
	imageModuleName = "image"
	webp            = "webp"
	resized         = "resized"
)

func newImageModule() *Module {
	m := &Module{
		Name:                     imageModuleName,
		Resolvers:                make(map[string]ResolverFunc),
		Defaults:                 make(Defaults),
		AllowedResolverArguments: make(map[string][]string),
	}

	//Set resolvers
	m.Resolvers[webp] = webpfn
	m.Resolvers[resized] = resizedfn

	//Set defaults
	m.Defaults[webp] = FalseStr
	m.Defaults[resized] = FalseStr

	m.AllowedResolverArguments[webp] = []string{TrueStr, FalseStr}
	m.AllowedResolverArguments[resized] = []string{TrueStr, FalseStr}

	return m
}

func webpfn(buff *bytes.Buffer, arg interface{}) error {

	if arg != TrueStr {
		return nil
	}

	img := bimg.NewImage(buff.Bytes())

	newimg, err := img.Convert(bimg.WEBP)
	if err != nil {
		return module_errors.WrapInternal(err, "image.webpfn.img.Convert")
	}

	(*buff).Reset()
	(*buff).Write(newimg)

	return nil
}

func resizedfn(buff *bytes.Buffer, arg interface{}) error {
	if arg != TrueStr {
		return nil
	}

	img := bimg.NewImage(buff.Bytes())

	size, err := img.Size()
	if err != nil {
		return module_errors.WrapInternal(err, "image.resizedfn.img.Size")
	}

	H := float64(size.Height)
	hmin := H * yAxisResizePercent
	hmax := H * (1 - 2*yAxisResizePercent)

	newimg, err := img.Extract(int(hmin), 0, size.Width, int(hmax))
	if err != nil {
		return module_errors.WrapInternal(err, "image.resizedfn.img.Extract")
	}

	(*buff).Reset()
	(*buff).Write(newimg)

	return nil
}
