package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	yamlv3 "gopkg.in/yaml.v3"

	diffutil "kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/third_party/dyff"
)

// supported diff-mode option values
const (
	ModeNormal      = "normal"
	ModeIgnoreAdded = "ignore-added"
	ModeLive        = "live"
)

type Options struct {
	swap                     bool
	translateListToDocuments bool
	fromLocation             string
	toLocation               string
	chroot                   string
	chrootFrom               string
	chrootTo                 string
	diffMode                 string
	sortByKubernetesResource bool
	outStyle                 string
	ignoreOrderChanges       bool
	noTableStyle             bool
	doNotInspectCerts        bool
	omitHeader               bool
	useGoPatchPaths          bool
	// exitWithCount      bool
}

func NewDiffOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	// diffed content from files
	switch {
	case len(args) == 1:
		o.fromLocation = args[0]
		o.toLocation = ""
	case len(args) == 2:
		o.fromLocation = args[0]
		o.toLocation = args[1]

		if o.swap {
			o.fromLocation, o.toLocation = o.toLocation, o.fromLocation
		}
	default:
		return fmt.Errorf("wrong number of args: %v,except 1 or 2", len(args))
	}

	// If the main change root flag is set, this (re-)sets the individual change roots of the two input files
	if o.chroot != "" {
		o.chrootFrom = o.chroot
		o.chrootTo = o.chroot
	}

	return nil
}

func (o *Options) Validate() error {
	switch strings.ToLower(o.diffMode) {
	case ModeLive:
	case ModeNormal:
	case ModeIgnoreAdded:
		break
	default:
		return fmt.Errorf("invalid diff mode `%s`", o.diffMode)
	}

	switch strings.ToLower(o.outStyle) {
	case diffutil.OutputHuman:
	case diffutil.OutputRaw:
		break
	default:
		return fmt.Errorf("invalid output style `%s`", o.outStyle)
	}

	return nil
}

func (o *Options) Run() error {
	var err error

	if strings.ToLower(o.diffMode) == ModeLive {
		if ytbx.IsStdin(o.fromLocation) {
			return liveDiffWithStdin()
		}

		return liveDiffWithFile(o.fromLocation, o.toLocation)
	}

	var from, to ytbx.InputFile

	if ytbx.IsStdin(o.fromLocation) {
		// diffed content from stdin
		result, err := ytbx.LoadFile(o.fromLocation)
		if err != nil {
			return wrap.Errorf(err, "failed to load file [%s]", o.fromLocation)
		}

		documents := result.Documents
		if len(documents) < 2 {
			return wrap.Errorf(err, "document size must greater than 1 when diffed content from stdin")
		}

		from = ytbx.InputFile{
			Location:  "-",
			Documents: []*yamlv3.Node{documents[0]},
		}
		to = ytbx.InputFile{
			Location:  "-",
			Documents: []*yamlv3.Node{documents[1]},
		}
	} else {
		from, to, err = ytbx.LoadFiles(o.fromLocation, o.toLocation)
		if err != nil {
			return wrap.Errorf(err, "failed to load input files")
		}
	}

	// Change root of 'from' input file if change root flag for 'from' is set
	if o.chrootFrom != "" {
		if err = dyff.ChangeRoot(&from, o.chrootFrom, o.useGoPatchPaths, o.translateListToDocuments); err != nil {
			return wrap.Errorf(err, "failed to change root of %s to path %s", from.Location, o.chrootFrom)
		}
	}

	// Change root of 'to' input file if change root flag for 'to' is set
	if o.chrootTo != "" {
		if err = dyff.ChangeRoot(&to, o.chrootTo, o.useGoPatchPaths, o.translateListToDocuments); err != nil {
			return wrap.Errorf(err, "failed to change root of %s to path %s", to.Location, o.chrootTo)
		}
	}

	// Sort 'to' input file by kubernetes kind order in 'from' input file
	if o.sortByKubernetesResource {
		if len(from.Documents) != len(to.Documents) {
			return wrap.Error(err, "from and to must have the same number of documents")
		}

		sort.Sort(k8sDocuments(from.Documents))
		sort.Sort(k8sDocuments(to.Documents))
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(o.ignoreOrderChanges))
	if err != nil {
		return wrap.Errorf(err, "failed to compare input files")
	}

	// handle diff-mode option
	switch strings.ToLower(o.diffMode) {
	case ModeNormal:
	case ModeIgnoreAdded:
		report = removeAddedElements(report)
	}

	// handle output option
	humanReport := &dyff.HumanReport{
		NoTableStyle:         o.noTableStyle,
		DoNotInspectCerts:    o.doNotInspectCerts,
		OmitHeader:           o.omitHeader,
		UseGoPatchPaths:      o.useGoPatchPaths,
		MinorChangeThreshold: 0.1,
		Report:               report,
	}
	reportString, err := diffutil.ToReportString(humanReport, o.outStyle)
	if err != nil {
		return err
	}
	fmt.Println(reportString)

	return nil
}
