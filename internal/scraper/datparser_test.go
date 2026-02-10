package scraper

import (
	"strings"
	"testing"
)

const sampleDAT = `<?xml version="1.0"?>
<datafile>
	<header>
		<name>Test DAT</name>
		<description>Test DAT file</description>
	</header>
	<game name="Super Mario Bros. (World)">
		<description>Super Mario Bros. (World)</description>
		<rom name="Super Mario Bros. (World).nes" size="40976"
			crc="3337EC46" md5="811b027eaf99c2def7b933c5208636de"
			sha1="facee9c577a5262dbe33b8370e8882c37ea48e2e"/>
	</game>
	<game name="Sonic the Hedgehog (USA, Europe)">
		<description>Sonic the Hedgehog (USA, Europe)</description>
		<rom name="Sonic the Hedgehog (USA, Europe).md" size="524288"
			crc="16FB1316" md5="1bc674be034e43c96b86487ac69d9293"
			sha1="26e4ee848d5b5ecd4af387eab571a0c3b6f2c4c8"/>
	</game>
</datafile>`

func TestParseDATReader(t *testing.T) {
	idx, err := ParseDATReader(strings.NewReader(sampleDAT))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if idx.Name != "Test DAT" {
		t.Errorf("expected name 'Test DAT', got %q", idx.Name)
	}

	if len(idx.ByCRC) != 2 {
		t.Errorf("expected 2 CRC entries, got %d", len(idx.ByCRC))
	}
}

func TestDATIndex_Lookup_ByCRC(t *testing.T) {
	idx, _ := ParseDATReader(strings.NewReader(sampleDAT))

	hashes := FileHashes{CRC32: "3337EC46"}
	entry, ok := idx.Lookup(hashes)
	if !ok {
		t.Fatal("expected to find entry by CRC")
	}
	if entry.GameName != "Super Mario Bros. (World)" {
		t.Errorf("unexpected game name: %s", entry.GameName)
	}
}

func TestDATIndex_Lookup_ByMD5(t *testing.T) {
	idx, _ := ParseDATReader(strings.NewReader(sampleDAT))

	hashes := FileHashes{MD5: "1bc674be034e43c96b86487ac69d9293"}
	entry, ok := idx.Lookup(hashes)
	if !ok {
		t.Fatal("expected to find entry by MD5")
	}
	if entry.GameName != "Sonic the Hedgehog (USA, Europe)" {
		t.Errorf("unexpected game name: %s", entry.GameName)
	}
}

func TestDATIndex_Lookup_BySHA1(t *testing.T) {
	idx, _ := ParseDATReader(strings.NewReader(sampleDAT))

	hashes := FileHashes{SHA1: "facee9c577a5262dbe33b8370e8882c37ea48e2e"}
	entry, ok := idx.Lookup(hashes)
	if !ok {
		t.Fatal("expected to find entry by SHA1")
	}
	if entry.GameName != "Super Mario Bros. (World)" {
		t.Errorf("unexpected game name: %s", entry.GameName)
	}
}

func TestDATIndex_Lookup_NotFound(t *testing.T) {
	idx, _ := ParseDATReader(strings.NewReader(sampleDAT))

	hashes := FileHashes{CRC32: "00000000"}
	_, ok := idx.Lookup(hashes)
	if ok {
		t.Error("expected no match")
	}
}

func TestDATIndex_Lookup_SHA1Priority(t *testing.T) {
	idx, _ := ParseDATReader(strings.NewReader(sampleDAT))

	// Provide both SHA1 and CRC - SHA1 should take priority
	hashes := FileHashes{
		SHA1:  "facee9c577a5262dbe33b8370e8882c37ea48e2e",
		CRC32: "16FB1316", // This is Sonic's CRC
	}
	entry, ok := idx.Lookup(hashes)
	if !ok {
		t.Fatal("expected match")
	}
	// SHA1 is Mario's, so we should get Mario
	if entry.GameName != "Super Mario Bros. (World)" {
		t.Errorf("expected SHA1 to take priority, got: %s", entry.GameName)
	}
}
