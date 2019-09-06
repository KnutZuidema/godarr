package torrent

import (
	"github.com/KnutZuidema/go-torrent/bencoding"
)

// Metainfo files (also known as .torrent files) are bencoded dictionaries
type Metainfo struct {
	// The URL of the tracker.
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
	// this is an extension to the official specification, offering backwards-compatibility.
	AnnounceList [][]string `bencode:"announce-list,omitempty"`
	// the creation time of the torrent, in standard UNIX epoch format (seconds since 1-Jan-1970 00:00:00 UTC)
	CreationDate int `bencode:"creation date,omitempty"`
	// free-form textual comments of the author
	Comment string `bencode:"comment,omitempty"`
	// name and version of the program used to create the .torrent
	CreatedBy string `bencode:"created by,omitempty"`
	// the string encoding format used to generate the pieces part of the info dictionary in the .torrent info
	Encoding string `bencode:"encoding,omitempty"`
}

// Info is a dictionary with information about a torrent file. There is a key length or a key files, but not both or
// neither. If length is present then the download represents a single file, otherwise it represents a set of files
// which go in a directory structure. For the purposes of the other keys, the multi-file case is treated as only having
// a single file by concatenating the files in the order they appear in the files list.
type Info struct {
	Files []File `bencode:"files,omitempty"`
	// the length of the file in bytes.
	Length int `bencode:"length,omitempty"`
	// a 32-character hexadecimal string corresponding to the MD5 sum of the file. This is not used by BitTorrent at
	// all, but it is included by some programs for greater compatibility.
	MD5Sum []byte `bencode:"md5sum"`
	// a UTF-8 encoded string which is the suggested name to save the file (or directory) as. It is purely advisory.
	// In the single file case, the name key is the name of a file, in the multiple file case, it's the name of a
	// directory.
	Name string `bencode:"name"`
	// the number of bytes in each piece the file is split into. For the purposes of transfer, files are split into
	// fixed-size pieces which are all the same length except for possibly the last one which may be truncated.
	// piece length is almost always a power of two, most commonly 2^18 = 256 K
	// (BitTorrent prior to version 3.2 uses 2^20 = 1 M as default).
	PieceLength int `bencode:"piece length"`
	// a string whose length is a multiple of 20. It is to be subdivided into strings of length 20, each of which is
	// the SHA1 hash of the piece at the corresponding index.
	Pieces []byte `bencode:"pieces"`
	// If it is set to true, the client MUST publish its presence to get other peers ONLY via the trackers explicitly
	// described in the metainfo file. If this field is set to false or is not present, the client may obtain peer from
	// other means, e.g. PEX peer exchange, dht. Here, "private" may be read as "no external peer source".
	Private bool `bencode:"private"`
}

type File struct {
	// The length of the file, in bytes.
	Length int `bencode:"length"`
	// a 32-character hexadecimal string corresponding to the MD5 sum of the file. This is not used by BitTorrent at
	// all, but it is included by some programs for greater compatibility.
	MD5Sum []byte `bencode:"md5sum"`
	// A list of UTF-8 encoded strings corresponding to subdirectory names, the last of which is the actual file name
	// (a zero length list is an error case).
	Path []string `bencode:"path"`
}

func NewMetainfoFromBytes(data []byte) (*Metainfo, error) {
	var info Metainfo
	if err := bencoding.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
