package nntputil

import (
	"testing"
	"time"

	"github.com/hobeone/gonab/db"
	"github.com/hobeone/gonab/types"
	"github.com/hobeone/nntp"

	. "github.com/onsi/gomega"
)

type FakeNNTPConnection struct {
	GroupResponse    *nntp.Group
	OverviewResponse []nntp.MessageOverview
}

func (f *FakeNNTPConnection) Group(g string) (*nntp.Group, error) {
	return f.GroupResponse, nil
}

func (f *FakeNNTPConnection) Overview(begin, end int64) ([]nntp.MessageOverview, error) {
	return f.OverviewResponse, nil
}

func TestRegexp(t *testing.T) {
	RegisterTestingT(t)

	s := `[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE] [16/16] - "[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE].mkv.vol63+29.par2" yEnc (27/30)`
	overview := nntp.MessageOverview{
		MessageID: "<foobaz12345@foo.bar>",
		Subject:   s,
		Bytes:     1024,
		Lines:     2,
		From:      "<foobaz@bar.com>",
		Date:      time.Now(),
		Extra:     []string{"Xref: number.nntp.giganews.com alt.binaries.multimedia.anime.highspeed:262328555"},
	}

	groupName := "misc.test"
	parts := overviewToParts(groupName, []nntp.MessageOverview{overview})

	Expect(parts).Should(HaveLen(1))
	Expect(parts).To(HaveKey("a4735742304a441"))
	part := parts["a4735742304a441"]
	Expect(part.Subject).To(Equal(`[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE] [16/16] - "[AnimeRG-FTS] Ajin (2016) - 02 [720p] [31FBC4AE].mkv.vol63+29.par2" yEnc`))
	Expect(part.TotalSegments).To(Equal(30))
	Expect(part.Segments).To(HaveLen(1))
	Expect(part.Segments[0].Segment).To(Equal(27))
}

func TestFindMissingMessages(t *testing.T) {
	RegisterTestingT(t)
	overviews := []nntp.MessageOverview{}
	missed := findMissingMessages(1, 10, overviews)
	Expect(missed).To(HaveLen(10))
	Expect(missed.Contains(types.MessageNumber(1))).To(BeTrue())
	Expect(missed.Contains(types.MessageNumber(10))).To(BeTrue())
}

func TestGroupScanForward(t *testing.T) {
	RegisterTestingT(t)

	fake := &FakeNNTPConnection{}
	nc := NewClient(fake)

	groupName := "alt.binaries.multimedia.anime"
	dbh := db.NewMemoryDBHandle(false)
	g := types.Group{
		Name:   groupName,
		Active: true,
	}
	fake.GroupResponse = &nntp.Group{
		Name:  groupName,
		High:  1000,
		Low:   100,
		Count: 900,
	}

	dbh.DB.Save(&g)

	err := nc.GroupScanForward(dbh, groupName, 100)
	Expect(err).To(BeNil())

	var missedCount int
	dbh.DB.Model(&types.MissedMessage{}).Count(&missedCount)
	Expect(missedCount).To(Equal(100))

	fake.OverviewResponse = []nntp.MessageOverview{
		{
			MessageNumber: 901,
			Subject:       "Subject Foo Yenc (1/30)",
			From:          "<foo@baz.bar>",
			Date:          time.Now(),
			MessageID:     "foo123456789@bar.com",
			Bytes:         12345,
			Lines:         1024,
			Extra: []string{
				"Xref: news.usenetserver.foo",
			},
		},
	}
	g.Last = 900
	dbh.DB.Save(&g)
	err = nc.GroupScanForward(dbh, groupName, 100)
	Expect(err).To(BeNil())

	dbh.DB.Model(&types.MissedMessage{}).Count(&missedCount)
	Expect(missedCount).To(Equal(100))

	var partCount int
	dbh.DB.Model(&types.Part{}).Count(&partCount)
	Expect(partCount).To(Equal(1))
}

// Faker that counts the number of times Overview was called
type FakeOverviewCounter struct {
	FakeNNTPConnection
	OverviewCalls int
}

func (f *FakeOverviewCounter) Overview(begin, end int64) ([]nntp.MessageOverview, error) {
	f.OverviewCalls++
	return f.OverviewResponse, nil
}

func TestGroupForwardScanSteps(t *testing.T) {
	RegisterTestingT(t)

	fake := &FakeOverviewCounter{}

	nc := NewClient(fake)
	nc.MaxScan = 100

	groupName := "alt.binaries.multimedia.anime"
	dbh := db.NewMemoryDBHandle(false)
	g := types.Group{
		Name:   groupName,
		Active: true,
		Last:   200,
		First:  100,
	}
	fake.GroupResponse = &nntp.Group{
		Name:  groupName,
		High:  2000,
		Low:   100,
		Count: 1900,
	}

	dbh.DB.Save(&g)

	err := nc.GroupScanForward(dbh, groupName, 1000)
	Expect(err).To(BeNil())

	Expect(fake.OverviewCalls).To(Equal(10))

	dbGroup, err := dbh.FindGroupByName(groupName)
	if err != nil {
		t.Fatalf("Error getting group %v", err)
	}
	Expect(dbGroup.Last).To(BeEquivalentTo(1200))
}

func BenchmarkHash(b *testing.B) {
	for n := 0; n < b.N; n++ {
		hashOverview("subject", "<from@bar.com>", "misc.test", 30)
	}
}
