package utils_test

import (
	"os"
	"path/filepath"

	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/operating"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpbackup/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("utils/filepath tests", func() {
	masterDir := "/data/gpseg-1"
	segDirOne := "/data/gpseg0"
	segDirTwo := "/data/gpseg1"
	var c *cluster.Cluster
	BeforeEach(func() {
		c = &cluster.Cluster{
			Segments: map[int]cluster.SegConfig{
				-1: {DataDir: masterDir},
			},
		}
	})
	Describe("Backup Filepath setup and accessors", func() {
		It("returns content dir for a single-host, single-segment nodes", func() {
			c.Segments[0] = cluster.SegConfig{DataDir: segDirOne}
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.SegDirMap).To(HaveLen(2))
			Expect(fpInfo.GetDirForContent(-1)).To(Equal("/data/gpseg-1/backups/20170101/20170101010101"))
			Expect(fpInfo.GetDirForContent(0)).To(Equal("/data/gpseg0/backups/20170101/20170101010101"))
		})
		It("sets up the configuration for a single-host, multi-segment fpInfo", func() {
			c.Segments[0] = cluster.SegConfig{DataDir: segDirOne}
			c.Segments[1] = cluster.SegConfig{DataDir: segDirTwo}
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.SegDirMap).To(HaveLen(3))
			Expect(fpInfo.GetDirForContent(-1)).To(Equal("/data/gpseg-1/backups/20170101/20170101010101"))
			Expect(fpInfo.GetDirForContent(0)).To(Equal("/data/gpseg0/backups/20170101/20170101010101"))
			Expect(fpInfo.GetDirForContent(1)).To(Equal("/data/gpseg1/backups/20170101/20170101010101"))
		})
		It("returns the content directory based on the user specified path", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.SegDirMap).To(HaveLen(1))
			Expect(fpInfo.GetDirForContent(-1)).To(Equal("/foo/bar/gpseg-1/backups/20170101/20170101010101"))
		})
	})
	Describe("GetTableBackupFilePathForCopyCommand()", func() {
		It("returns table file path for copy command", func() {
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePathForCopyCommand(1234, false)).To(Equal("<SEG_DATA_DIR>/backups/20170101/20170101010101/gpbackup_<SEGID>_20170101010101_1234"))
		})
		It("returns table file path for copy command based on user specified path", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePathForCopyCommand(1234, false)).To(Equal("/foo/bar/gpseg<SEGID>/backups/20170101/20170101010101/gpbackup_<SEGID>_20170101010101_1234"))
		})
		It("returns table file path for copy command in single-file mode", func() {
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePathForCopyCommand(1234, true)).To(Equal("<SEG_DATA_DIR>/backups/20170101/20170101010101/gpbackup_<SEGID>_20170101010101"))
		})
		It("returns table file path for copy command based on user specified path in single-file mode", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePathForCopyCommand(1234, true)).To(Equal("/foo/bar/gpseg<SEGID>/backups/20170101/20170101010101/gpbackup_<SEGID>_20170101010101"))
		})
	})
	Describe("GetReportFilePath", func() {
		It("returns report file path", func() {
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.GetBackupReportFilePath()).To(Equal("/data/gpseg-1/backups/20170101/20170101010101/gpbackup_20170101010101_report"))
		})
		It("returns report file path based on user specified path", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.GetBackupReportFilePath()).To(Equal("/foo/bar/gpseg-1/backups/20170101/20170101010101/gpbackup_20170101010101_report"))
		})
	})
	Describe("GetTableBackupFilePath", func() {
		It("returns table file path", func() {
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePath(-1, 1234, false)).To(Equal("/data/gpseg-1/backups/20170101/20170101010101/gpbackup_-1_20170101010101_1234"))
		})
		It("returns table file path based on user specified path", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePath(-1, 1234, false)).To(Equal("/foo/bar/gpseg-1/backups/20170101/20170101010101/gpbackup_-1_20170101010101_1234"))
		})
		It("returns single data file path", func() {
			fpInfo := utils.NewFilePathInfo(c, "", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePath(-1, 1234, true)).To(Equal("/data/gpseg-1/backups/20170101/20170101010101/gpbackup_-1_20170101010101"))
		})
		It("returns single data file path based on user specified path", func() {
			fpInfo := utils.NewFilePathInfo(c, "/foo/bar", "20170101010101", "gpseg")
			Expect(fpInfo.GetTableBackupFilePath(-1, 1234, true)).To(Equal("/foo/bar/gpseg-1/backups/20170101/20170101010101/gpbackup_-1_20170101010101"))
		})
	})
	Describe("ParseSegPrefix", func() {
		AfterEach(func() {
			operating.System.Glob = filepath.Glob
		})
		It("returns segment prefix from directory path if master backup directory exists", func() {
			operating.System.Glob = func(pattern string) (matches []string, err error) {
				return []string{"/tmp/foo/gpseg-1/backups/datestamp1/timestamp1"}, nil
			}

			Expect(utils.ParseSegPrefix("/tmp/foo", "timestamp1")).To(Equal("gpseg"))
		})
		It("returns empty string if backup directory is empty", func() {
			Expect(utils.ParseSegPrefix("", "timestamp1")).To(Equal(""))
		})
		It("panics if master backup directory does not exist", func() {
			operating.System.Glob = func(pattern string) (matches []string, err error) { return []string{}, nil }
			defer testhelper.ShouldPanicWithMessage("Master backup directory in /tmp/foo missing or inaccessible")
			Expect(utils.ParseSegPrefix("/tmp/foo", "timestamp1")).To(Equal("gpseg"))
		})
		It("panics if there is an error accessing master backup directory", func() {
			operating.System.Glob = func(pattern string) (matches []string, err error) {
				return []string{""}, os.ErrPermission
			}
			defer testhelper.ShouldPanicWithMessage("Master backup directory in /tmp/foo missing or inaccessible")
			Expect(utils.ParseSegPrefix("/tmp/foo", "timestamp1")).To(Equal("gpseg"))
		})
	})
})
