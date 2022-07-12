package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gonvenience/wrap"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	diffutil "kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/third_party/diff"
	"kusionstack.io/kusion/third_party/dyff"
)

var (
	diffShort = "Compare differences between input files <from> and <to>"

	diffLong = `
		Compare files differences and display the delta.
		Support input file types are: YAML (http://yaml.org/) and JSON (http://json.org/).`

	diffExample = `
		# The comparison object comes from the files
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# The comparison object comes from the stdin
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff -`
)

func NewCmdDiff() *cobra.Command {
	o := NewDiffOptions()

	cmd := &cobra.Command{
		Use:     "diff <from> <to>",
		Short:   i18n.T(diffShort),
		Long:    templates.LongDesc(i18n.T(diffLong)),
		Args:    cobra.RangeArgs(1, 2),
		Aliases: []string{"df"},
		Example: templates.Examples(i18n.T(diffExample)),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Complete(args)
			if err != nil {
				return err
			}
			err = o.Validate()
			if err != nil {
				return err
			}
			err = o.Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	// Input documents modification flags
	cmd.Flags().BoolVar(&o.swap, "swap", false,
		i18n.T("Swap <from> and <to> for comparison. Note that it is invalid when <from> is stdin. The default is false"))
	cmd.Flags().StringVar(&o.diffMode, "diff-mode", "normal",
		i18n.T(fmt.Sprintf("Diff mode. One of %s and %s. The default is normal", DiffModeNormal, DiffModeIgnoreAdded)))
	cmd.Flags().StringVarP(&o.outStyle, "output", "o", "human",
		i18n.T(fmt.Sprintf("Specify the output style. One of %s and %s. The default is human", diffutil.OutputHuman, diffutil.OutputRaw)))
	cmd.Flags().BoolVarP(&o.ignoreOrderChanges, "ignore-order-changes", "i", false,
		i18n.T("Ignore order changes in lists. The default is false"))
	cmd.Flags().BoolVarP(&o.omitHeader, "omit-header", "b", false,
		i18n.T("Omit the dyff summary header. The default is false"))
	cmd.Flags().BoolVarP(&o.sortByKubernetesResource, "sort-by-kubernetes-resource", "k", true,
		i18n.T("Sort from and to by kubernetes resource order(non standard behavior). The default is false"))

	return cmd
}

func liveDiffWithStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to load data from os.stdin as %v", err)
	}

	results := []*unstructured.Unstructured{}
	decoder := yamlv3.NewDecoder(bytes.NewReader(data))

	for {
		m := make(map[string]interface{})

		err := decoder.Decode(&m)
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to decode stdin content as %v", err)
		}

		result := &unstructured.Unstructured{}
		result.Object = m
		results = append(results, result)
	}

	ignorePaths := []string{
		"/metadata/annotations/helm.sh~1hook", "/metadata/annotations/helm.sh~1hook-weight",
		"/spec/strategy/groupSerials", "/spec/strategy/partitions", "/spec/replicas",
	}

	ignoreNormalizer, err := diff.NewDefaultIgnoreNormalizer(ignorePaths)
	if err != nil {
		return wrap.Errorf(err, "failed to create normalizer as %v", err)
	}

	diffRes, err := diff.Diff(results[0], results[1], diff.WithNormalizer(ignoreNormalizer))
	if err != nil {
		return wrap.Errorf(err, "failed to calculate diff as %v", err)
	}

	resultToOutput, err := json.Marshal(diffRes)
	if err != nil {
		return wrap.Errorf(err, "failed to marshal diff result as %v", err)
	}

	fmt.Println(string(resultToOutput))

	return nil
}

func liveDiffWithFile(fromLocation, toLocation string) error {
	type resultPair struct {
		result *unstructured.Unstructured
		err    error
	}

	fromChan := make(chan resultPair, 1)
	toChan := make(chan resultPair, 1)

	go func() {
		result, err := loadFile(fromLocation)
		fromChan <- resultPair{result, err}
	}()

	go func() {
		result, err := loadFile(toLocation)
		toChan <- resultPair{result, err}
	}()

	from := <-fromChan
	if from.err != nil {
		return from.err
	}

	to := <-toChan
	if to.err != nil {
		return to.err
	}

	ignorePaths := []string{
		"/metadata/annotations/helm.sh~1hook", "/metadata/annotations/helm.sh~1hook-weight",
		"/spec/strategy/groupSerials", "/spec/strategy/partitions", "/spec/replicas",
	}

	ignoreNormalizer, err := diff.NewDefaultIgnoreNormalizer(ignorePaths)
	if err != nil {
		return wrap.Errorf(err, "failed to create normalizer as %v", err)
	}

	diffRes, err := diff.Diff(from.result, to.result, diff.WithNormalizer(ignoreNormalizer))
	if err != nil {
		return wrap.Errorf(err, "failed to calculate diff as %v", err)
	}

	resultToOutput, err := json.Marshal(diffRes)
	if err != nil {
		return wrap.Errorf(err, "failed to marshal diff result as %v", err)
	}

	fmt.Println(string(resultToOutput))

	return nil
}

func loadFile(location string) (*unstructured.Unstructured, error) {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, wrap.Errorf(err, "failed to load input files")
	}

	result := &unstructured.Unstructured{}

	err = yaml.Unmarshal(data, result)
	if err != nil {
		return nil, wrap.Errorf(err, "failed to unmarshal target state as %v", err)
	}

	return result, nil
}

func removeAddedElements(report dyff.Report) dyff.Report {
	newDiffList := make([]dyff.Diff, 0, len(report.Diffs))

	for _, diff := range report.Diffs {
		newDiff := dyff.Diff{
			Path:    diff.Path,
			Details: make([]dyff.Detail, 0, len(diff.Details)),
		}
		// remove element of kind is `+`
		for _, detail := range diff.Details {
			if detail.Kind != '+' {
				newDiff.Details = append(newDiff.Details, detail)
			}
		}
		// remove element of detail length is 0
		if len(newDiff.Details) != 0 {
			newDiffList = append(newDiffList, newDiff)
		}
	}

	report.Diffs = newDiffList

	return report
}
