package get_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get healthchecks", func() {

	var sampleHealthChecks []*core_mesh.HealthCheckResource
	BeforeEach(func() {
		sampleHealthChecks = []*core_mesh.HealthCheckResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "web-to-backend",
				},
				Spec: &mesh_proto.HealthCheck{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "backend-to-db",
				},
				Spec: &mesh_proto.HealthCheck{},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "gateway-to-service",
				},
				Spec: &mesh_proto.HealthCheck{},
			},
		}
	})

	Describe("GetHealthChecksCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, core_mesh.HealthCheckResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, pt := range sampleHealthChecks {
				key := core_model.ResourceKey{
					Mesh: pt.Meta.GetMesh(),
					Name: pt.Meta.GetName(),
				}
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			pagination   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get healthchecks -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "healthchecks", given.outputFormat, given.pagination),
				).To(Succeed())

				// when
				expected, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-healthchecks.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-healthchecks.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-healthchecks.pagination.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-healthchecks.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-healthchecks.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
