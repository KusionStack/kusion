package diff

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	yamlv3 "gopkg.in/yaml.v3"
	"kusionstack.io/kusion/third_party/dyff"
)

// supported diff-mode option values
const (
	DiffModeNormal      = "normal"
	DiffModeIgnoreAdded = "ignore-added"
	DiffModeLive        = "live"
)

type DiffOptions struct {
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

func NewDiffOptions() *DiffOptions {
	return &DiffOptions{}
}

func (o *DiffOptions) Complete(args []string) error {
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

func (o *DiffOptions) Validate() error {
	switch strings.ToLower(o.diffMode) {
	case DiffModeLive:
	case DiffModeNormal:
	case DiffModeIgnoreAdded:
		break
	default:
		return fmt.Errorf("invalid diff mode `%s`", o.diffMode)
	}

	switch strings.ToLower(o.outStyle) {
	case OutputHuman:
	case OutputRaw:
		break
	default:
		return fmt.Errorf("invalid output style `%s`", o.outStyle)
	}

	return nil
}

func (o *DiffOptions) Run() error {
	var err error

	if strings.ToLower(o.diffMode) == DiffModeLive {
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
	case DiffModeNormal:
	case DiffModeIgnoreAdded:
		report = removeAddedElements(report)
	}

	// handle output option
	switch strings.ToLower(o.outStyle) {
	case OutputHuman:
		return o.writeReport(report)
	case OutputRaw:
		// output stdout/file
		reportMap := map[string]interface{}{
			"diffs": report.Diffs,
		}

		reportYAML, err := yamlv3.Marshal(reportMap)
		if err != nil {
			return wrap.Errorf(err, "failed to marshal report diffs")
		}

		fmt.Println(string(reportYAML))

		return nil
	}

	return nil
}

func (o *DiffOptions) writeReport(report dyff.Report) error {
	reportWriter := &dyff.HumanReport{
		Report:               report,
		DoNotInspectCerts:    o.doNotInspectCerts,
		NoTableStyle:         o.noTableStyle,
		OmitHeader:           o.omitHeader,
		UseGoPatchPaths:      o.useGoPatchPaths,
		MinorChangeThreshold: 0.1,
	}

	if err := reportWriter.WriteReport(os.Stdout); err != nil {
		return wrap.Errorf(err, "failed to print report")
	}

	return nil
}
