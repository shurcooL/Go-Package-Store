// +build ignore

package main

import (
	"bytes"
	"go/build"
	"go/token"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/shurcooL/Go-Package-Store"

	. "gist.github.com/5286084.git"

	"gist.github.com/7480523.git"
	"github.com/google/go-github/github"
)

func NewString(s string) *string {
	return &s
}

func NewInt(i int) *int {
	return &i
}

func main() {
	buf := new(bytes.Buffer)

	goPackage := (*gist7480523.GoPackage)(&gist7480523.GoPackage{
		Bpkg: (*build.Package)(&build.Package{
			Dir:         (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml"),
			Name:        (string)("toml"),
			Doc:         (string)("Package toml provides facilities for decoding TOML configuration files via reflection."),
			ImportPath:  (string)("github.com/BurntSushi/toml"),
			Root:        (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding"),
			SrcRoot:     (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src"),
			PkgRoot:     (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/pkg"),
			BinDir:      (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/bin"),
			Goroot:      (bool)(false),
			PkgObj:      (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/pkg/darwin_amd64/github.com/BurntSushi/toml.a"),
			AllTags:     ([]string)([]string{}),
			ConflictDir: (string)(""),
			GoFiles: ([]string)([]string{
				(string)("decode.go"),
				(string)("doc.go"),
				(string)("encode.go"),
				(string)("lex.go"),
				(string)("parse.go"),
				(string)("type_check.go"),
				(string)("type_fields.go"),
			}),
			CgoFiles:       ([]string)([]string{}),
			IgnoredGoFiles: ([]string)([]string{}),
			CFiles:         ([]string)([]string{}),
			CXXFiles:       ([]string)([]string{}),
			HFiles:         ([]string)([]string{}),
			SFiles:         ([]string)([]string{}),
			SwigFiles:      ([]string)([]string{}),
			SwigCXXFiles:   ([]string)([]string{}),
			SysoFiles:      ([]string)([]string{}),
			CgoCFLAGS:      ([]string)([]string{}),
			CgoCPPFLAGS:    ([]string)([]string{}),
			CgoCXXFLAGS:    ([]string)([]string{}),
			CgoLDFLAGS:     ([]string)([]string{}),
			CgoPkgConfig:   ([]string)([]string{}),
			Imports: ([]string)([]string{
				(string)("bufio"),
				(string)("encoding"),
				(string)("errors"),
				(string)("fmt"),
				(string)("io"),
				(string)("io/ioutil"),
				(string)("log"),
				(string)("reflect"),
				(string)("sort"),
				(string)("strconv"),
				(string)("strings"),
				(string)("sync"),
				(string)("time"),
				(string)("unicode/utf8"),
			}),
			ImportPos: (map[string][]token.Position)(map[string][]token.Position{
				(string)("reflect"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(62),
						Line:     (int)(8),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(700),
						Line:     (int)(22),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
						Offset:   (int)(252),
						Line:     (int)(10),
						Column:   (int)(2),
					}),
				}),
				(string)("sort"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(711),
						Line:     (int)(23),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
						Offset:   (int)(263),
						Line:     (int)(11),
						Column:   (int)(2),
					}),
				}),
				(string)("log"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(31),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
				}),
				(string)("fmt"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(36),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(687),
						Line:     (int)(20),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
				(string)("io/ioutil"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(49),
						Line:     (int)(7),
						Column:   (int)(2),
					}),
				}),
				(string)("strings"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(73),
						Line:     (int)(9),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(730),
						Line:     (int)(25),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(49),
						Line:     (int)(7),
						Column:   (int)(2),
					}),
				}),
				(string)("bufio"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(656),
						Line:     (int)(17),
						Column:   (int)(2),
					}),
				}),
				(string)("encoding"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(665),
						Line:     (int)(18),
						Column:   (int)(2),
					}),
				}),
				(string)("time"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(84),
						Line:     (int)(10),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(60),
						Line:     (int)(8),
						Column:   (int)(2),
					}),
				}),
				(string)("errors"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(677),
						Line:     (int)(19),
						Column:   (int)(2),
					}),
				}),
				(string)("strconv"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(719),
						Line:     (int)(24),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(38),
						Line:     (int)(6),
						Column:   (int)(2),
					}),
				}),
				(string)("unicode/utf8"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex.go"),
						Offset:   (int)(31),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
						Offset:   (int)(68),
						Line:     (int)(9),
						Column:   (int)(2),
					}),
				}),
				(string)("sync"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
						Offset:   (int)(271),
						Line:     (int)(12),
						Column:   (int)(2),
					}),
				}),
				(string)("io"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
						Offset:   (int)(43),
						Line:     (int)(6),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
						Offset:   (int)(694),
						Line:     (int)(21),
						Column:   (int)(2),
					}),
				}),
			}),
			TestGoFiles: ([]string)([]string{
				(string)("decode_test.go"),
				(string)("encode_test.go"),
				(string)("lex_test.go"),
				(string)("out_test.go"),
				(string)("parse_test.go"),
			}),
			TestImports: ([]string)([]string{
				(string)("bytes"),
				(string)("encoding/json"),
				(string)("flag"),
				(string)("fmt"),
				(string)("log"),
				(string)("reflect"),
				(string)("strings"),
				(string)("testing"),
				(string)("time"),
			}),
			TestImportPos: (map[string][]token.Position)(map[string][]token.Position{
				(string)("fmt"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(41),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/out_test.go"),
						Offset:   (int)(32),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
				}),
				(string)("log"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(48),
						Line:     (int)(6),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex_test.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
				(string)("reflect"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(55),
						Line:     (int)(7),
						Column:   (int)(2),
					}),
				}),
				(string)("bytes"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode_test.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
				(string)("encoding/json"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
				(string)("testing"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(66),
						Line:     (int)(8),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode_test.go"),
						Offset:   (int)(33),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex_test.go"),
						Offset:   (int)(31),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse_test.go"),
						Offset:   (int)(35),
						Line:     (int)(5),
						Column:   (int)(2),
					}),
				}),
				(string)("time"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
						Offset:   (int)(77),
						Line:     (int)(9),
						Column:   (int)(2),
					}),
				}),
				(string)("flag"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/out_test.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
				(string)("strings"): ([]token.Position)([]token.Position{
					(token.Position)(token.Position{
						Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse_test.go"),
						Offset:   (int)(24),
						Line:     (int)(4),
						Column:   (int)(2),
					}),
				}),
			}),
			XTestGoFiles:   ([]string)([]string{}),
			XTestImports:   ([]string)([]string{}),
			XTestImportPos: (map[string][]token.Position)(map[string][]token.Position{}),
		}),
		Standard:    (bool)(false),
		Vcs:         (nil),
		Status:      (string)(""),
		LocalBranch: (string)("master"),
		Local:       (string)("d7b4e27ae7df432264ca4ecf2dbec313ed01c330"),
		Remote:      (string)("f8260fb5e94dba7ed68a2621b5c4fdc675bd3861"),
	})
	cc := (*github.CommitsComparison)(&github.CommitsComparison{
		BaseCommit: (*github.RepositoryCommit)(&github.RepositoryCommit{
			SHA: (*string)(NewString("d7b4e27ae7df432264ca4ecf2dbec313ed01c330")),
			Commit: (*github.Commit)(&github.Commit{
				SHA: (*string)(nil),
				Author: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Andrew Gallant")),
					Email: (*string)(NewString("jamslam@gmail.com")),
				}),
				Committer: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Andrew Gallant")),
					Email: (*string)(NewString("jamslam@gmail.com")),
				}),
				Message: (*string)(NewString("Merge pull request #16 from nobonobo/master\n\nInfinite loop avoidance in Unexpected EOF")),
				Tree: (*github.Tree)(&github.Tree{
					SHA:     (*string)(NewString("7b938c31378d4b37c244f66d62400d8b3e44bfdd")),
					Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
				}),
				Parents: ([]github.Commit)([]github.Commit{}),
				Stats:   (*github.CommitStats)(nil),
			}),
			Author: (*github.User)(&github.User{
				Login:       (*string)(NewString("BurntSushi")),
				ID:          (*int)(NewInt(456674)),
				URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
				GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*time.Time)(nil),
			}),
			Committer: (*github.User)(&github.User{
				Login:       (*string)(NewString("BurntSushi")),
				ID:          (*int)(NewInt(456674)),
				URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
				GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*time.Time)(nil),
			}),
			Parents: ([]github.Commit)([]github.Commit{
				(github.Commit)(github.Commit{
					SHA:       (*string)(NewString("2fffd0e6ca4b88558be4bcab497231c95270cd07")),
					Author:    (*github.CommitAuthor)(nil),
					Committer: (*github.CommitAuthor)(nil),
					Message:   (*string)(nil),
					Tree:      (*github.Tree)(nil),
					Parents:   ([]github.Commit)([]github.Commit{}),
					Stats:     (*github.CommitStats)(nil),
				}),
				(github.Commit)(github.Commit{
					SHA:       (*string)(NewString("ff98ae77642e0bf7f0e2b63857903f44d88f5b5e")),
					Author:    (*github.CommitAuthor)(nil),
					Committer: (*github.CommitAuthor)(nil),
					Message:   (*string)(nil),
					Tree:      (*github.Tree)(nil),
					Parents:   ([]github.Commit)([]github.Commit{}),
					Stats:     (*github.CommitStats)(nil),
				}),
			}),
			Message: (*string)(nil),
			Stats:   (*github.CommitStats)(nil),
			Files:   ([]github.CommitFile)([]github.CommitFile{}),
		}),
		Status:       (*string)(NewString("ahead")),
		AheadBy:      (*int)(NewInt(3)),
		BehindBy:     (*int)(NewInt(0)),
		TotalCommits: (*int)(NewInt(3)),
		Commits: ([]github.RepositoryCommit)([]github.RepositoryCommit{
			(github.RepositoryCommit)(github.RepositoryCommit{
				SHA: (*string)(NewString("629e931d4930dcd3dc393b700a6d4dcd487441b0")),
				Commit: (*github.Commit)(&github.Commit{
					SHA: (*string)(nil),
					Author: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Rafal Jeczalik")),
						Email: (*string)(NewString("rjeczalik@gmail.com")),
					}),
					Committer: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Rafal Jeczalik")),
						Email: (*string)(NewString("rjeczalik@gmail.com")),
					}),
					Message: (*string)(NewString("gofmt")),
					Tree: (*github.Tree)(&github.Tree{
						SHA:     (*string)(NewString("858831a3d12594b093954d3b27df62bd57e76b5f")),
						Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
					}),
					Parents: ([]github.Commit)([]github.Commit{}),
					Stats:   (*github.CommitStats)(nil),
				}),
				Author: (*github.User)(&github.User{
					Login:       (*string)(NewString("rjeczalik")),
					ID:          (*int)(NewInt(1162017)),
					URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
					GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Committer: (*github.User)(&github.User{
					Login:       (*string)(NewString("rjeczalik")),
					ID:          (*int)(NewInt(1162017)),
					URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
					GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Parents: ([]github.Commit)([]github.Commit{
					(github.Commit)(github.Commit{
						SHA:       (*string)(NewString("d7b4e27ae7df432264ca4ecf2dbec313ed01c330")),
						Author:    (*github.CommitAuthor)(nil),
						Committer: (*github.CommitAuthor)(nil),
						Message:   (*string)(nil),
						Tree:      (*github.Tree)(nil),
						Parents:   ([]github.Commit)([]github.Commit{}),
						Stats:     (*github.CommitStats)(nil),
					}),
				}),
				Message: (*string)(nil),
				Stats:   (*github.CommitStats)(nil),
				Files:   ([]github.CommitFile)([]github.CommitFile{}),
			}),
			(github.RepositoryCommit)(github.RepositoryCommit{
				SHA: (*string)(NewString("6cab9f41ecc899af473584dbeff6e1814a098a6c")),
				Commit: (*github.Commit)(&github.Commit{
					SHA: (*string)(nil),
					Author: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Rafal Jeczalik")),
						Email: (*string)(NewString("rjeczalik@gmail.com")),
					}),
					Committer: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Rafal Jeczalik")),
						Email: (*string)(NewString("rjeczalik@gmail.com")),
					}),
					Message: (*string)(NewString("fix go vet warnings")),
					Tree: (*github.Tree)(&github.Tree{
						SHA:     (*string)(NewString("896b4c18dcc467cd0a58c3d3d71300849eea68b8")),
						Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
					}),
					Parents: ([]github.Commit)([]github.Commit{}),
					Stats:   (*github.CommitStats)(nil),
				}),
				Author: (*github.User)(&github.User{
					Login:       (*string)(NewString("rjeczalik")),
					ID:          (*int)(NewInt(1162017)),
					URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
					GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Committer: (*github.User)(&github.User{
					Login:       (*string)(NewString("rjeczalik")),
					ID:          (*int)(NewInt(1162017)),
					URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
					GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Parents: ([]github.Commit)([]github.Commit{
					(github.Commit)(github.Commit{
						SHA:       (*string)(NewString("629e931d4930dcd3dc393b700a6d4dcd487441b0")),
						Author:    (*github.CommitAuthor)(nil),
						Committer: (*github.CommitAuthor)(nil),
						Message:   (*string)(nil),
						Tree:      (*github.Tree)(nil),
						Parents:   ([]github.Commit)([]github.Commit{}),
						Stats:     (*github.CommitStats)(nil),
					}),
				}),
				Message: (*string)(nil),
				Stats:   (*github.CommitStats)(nil),
				Files:   ([]github.CommitFile)([]github.CommitFile{}),
			}),
			(github.RepositoryCommit)(github.RepositoryCommit{
				SHA: (*string)(NewString("f8260fb5e94dba7ed68a2621b5c4fdc675bd3861")),
				Commit: (*github.Commit)(&github.Commit{
					SHA: (*string)(nil),
					Author: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Andrew Gallant")),
						Email: (*string)(NewString("jamslam@gmail.com")),
					}),
					Committer: (*github.CommitAuthor)(&github.CommitAuthor{
						Date:  (*time.Time)(nil),
						Name:  (*string)(NewString("Andrew Gallant")),
						Email: (*string)(NewString("jamslam@gmail.com")),
					}),
					Message: (*string)(NewString("We want %s since errorf escapes some characters (like new lines), which turns them into strings.")),
					Tree: (*github.Tree)(&github.Tree{
						SHA:     (*string)(NewString("94a352d78ef7c5484d13f43663492e137988627b")),
						Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
					}),
					Parents: ([]github.Commit)([]github.Commit{}),
					Stats:   (*github.CommitStats)(nil),
				}),
				Author: (*github.User)(&github.User{
					Login:       (*string)(NewString("BurntSushi")),
					ID:          (*int)(NewInt(456674)),
					URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
					GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Committer: (*github.User)(&github.User{
					Login:       (*string)(NewString("BurntSushi")),
					ID:          (*int)(NewInt(456674)),
					URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
					AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
					GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
					Name:        (*string)(nil),
					Company:     (*string)(nil),
					Blog:        (*string)(nil),
					Location:    (*string)(nil),
					Email:       (*string)(nil),
					Hireable:    (*bool)(nil),
					PublicRepos: (*int)(nil),
					Followers:   (*int)(nil),
					Following:   (*int)(nil),
					CreatedAt:   (*time.Time)(nil),
				}),
				Parents: ([]github.Commit)([]github.Commit{
					(github.Commit)(github.Commit{
						SHA:       (*string)(NewString("6cab9f41ecc899af473584dbeff6e1814a098a6c")),
						Author:    (*github.CommitAuthor)(nil),
						Committer: (*github.CommitAuthor)(nil),
						Message:   (*string)(nil),
						Tree:      (*github.Tree)(nil),
						Parents:   ([]github.Commit)([]github.Commit{}),
						Stats:     (*github.CommitStats)(nil),
					}),
				}),
				Message: (*string)(nil),
				Stats:   (*github.CommitStats)(nil),
				Files:   ([]github.CommitFile)([]github.CommitFile{}),
			}),
		}),
		Files: ([]github.CommitFile)([]github.CommitFile{
			(github.CommitFile)(github.CommitFile{
				SHA:       (*string)(NewString("27cea0fdf82c428519b9dbbd67df183853720c97")),
				Filename:  (*string)(NewString("encode_test.go")),
				Additions: (*int)(NewInt(14)),
				Deletions: (*int)(NewInt(14)),
				Changes:   (*int)(NewInt(28)),
				Status:    (*string)(NewString("modified")),
				Patch:     (*string)(NewString("@@ -75,29 +75,29 @@ func TestEncode(t *testing.T) {\n \t\t\t\tSliceOfMixedArrays    [][2]interface{}\n \t\t\t\tArrayOfMixedSlices    [2][]interface{}\n \t\t\t}{\n-\t\t\t\t[][2]int{[2]int{1, 2}, [2]int{3, 4}},\n-\t\t\t\t[2][]int{[]int{1, 2}, []int{3, 4}},\n+\t\t\t\t[][2]int{{1, 2}, {3, 4}},\n+\t\t\t\t[2][]int{{1, 2}, {3, 4}},\n \t\t\t\t[][2][]int{\n-\t\t\t\t\t[2][]int{\n-\t\t\t\t\t\t[]int{1, 2}, []int{3, 4},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{1, 2}, {3, 4},\n \t\t\t\t\t},\n-\t\t\t\t\t[2][]int{\n-\t\t\t\t\t\t[]int{5, 6}, []int{7, 8},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{5, 6}, {7, 8},\n \t\t\t\t\t},\n \t\t\t\t},\n \t\t\t\t[2][][2]int{\n-\t\t\t\t\t[][2]int{\n-\t\t\t\t\t\t[2]int{1, 2}, [2]int{3, 4},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{1, 2}, {3, 4},\n \t\t\t\t\t},\n-\t\t\t\t\t[][2]int{\n-\t\t\t\t\t\t[2]int{5, 6}, [2]int{7, 8},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{5, 6}, {7, 8},\n \t\t\t\t\t},\n \t\t\t\t},\n \t\t\t\t[][2]interface{}{\n-\t\t\t\t\t[2]interface{}{1, 2}, [2]interface{}{\"a\", \"b\"},\n+\t\t\t\t\t{1, 2}, {\"a\", \"b\"},\n \t\t\t\t},\n \t\t\t\t[2][]interface{}{\n-\t\t\t\t\t[]interface{}{1, 2}, []interface{}{\"a\", \"b\"},\n+\t\t\t\t\t{1, 2}, {\"a\", \"b\"},\n \t\t\t\t},\n \t\t\t},\n \t\t\twantOutput: `SliceOfArrays = [[1, 2], [3, 4]]\n@@ -162,8 +162,8 @@ ArrayOfMixedSlices = [[1, 2], [\"a\", \"b\"]]`,\n \t\t},\n \t\t\"nested map\": {\n \t\t\tinput: map[string]map[string]int{\n-\t\t\t\t\"a\": map[string]int{\"b\": 1},\n-\t\t\t\t\"c\": map[string]int{\"d\": 2},\n+\t\t\t\t\"a\": {\"b\": 1},\n+\t\t\t\t\"c\": {\"d\": 2},\n \t\t\t},\n \t\t\twantOutput: \"[a]\\n  b = 1\\n\\n[c]\\n  d = 2\",\n \t\t},")),
			}),
			(github.CommitFile)(github.CommitFile{
				SHA:       (*string)(NewString("43afe3c3fda0a46e22d0d66620f61c22e2e6a57e")),
				Filename:  (*string)(NewString("parse.go")),
				Additions: (*int)(NewInt(8)),
				Deletions: (*int)(NewInt(8)),
				Changes:   (*int)(NewInt(16)),
				Status:    (*string)(NewString("modified")),
				Patch:     (*string)(NewString("@@ -65,7 +65,7 @@ func parse(data string) (p *parser, err error) {\n \treturn p, nil\n }\n \n-func (p *parser) panic(format string, v ...interface{}) {\n+func (p *parser) panicf(format string, v ...interface{}) {\n \tmsg := fmt.Sprintf(\"Near line %d, key '%s': %s\",\n \t\tp.approxLine, p.current(), fmt.Sprintf(format, v...))\n \tpanic(parseError(msg))\n@@ -74,7 +74,7 @@ func (p *parser) panic(format string, v ...interface{}) {\n func (p *parser) next() item {\n \tit := p.lx.nextItem()\n \tif it.typ == itemError {\n-\t\tp.panic(\"Near line %d: %s\", it.line, it.val)\n+\t\tp.panicf(\"Near line %d: %s\", it.line, it.val)\n \t}\n \treturn it\n }\n@@ -164,7 +164,7 @@ func (p *parser) value(it item) (interface{}, tomlType) {\n \t\t\tif e, ok := err.(*strconv.NumError); ok &&\n \t\t\t\te.Err == strconv.ErrRange {\n \n-\t\t\t\tp.panic(\"Integer '%s' is out of the range of 64-bit \"+\n+\t\t\t\tp.panicf(\"Integer '%s' is out of the range of 64-bit \"+\n \t\t\t\t\t\"signed integers.\", it.val)\n \t\t\t} else {\n \t\t\t\tp.bug(\"Expected integer value, but got '%s'.\", it.val)\n@@ -184,7 +184,7 @@ func (p *parser) value(it item) (interface{}, tomlType) {\n \t\t\tif e, ok := err.(*strconv.NumError); ok &&\n \t\t\t\te.Err == strconv.ErrRange {\n \n-\t\t\t\tp.panic(\"Float '%s' is out of the range of 64-bit \"+\n+\t\t\t\tp.panicf(\"Float '%s' is out of the range of 64-bit \"+\n \t\t\t\t\t\"IEEE-754 floating-point numbers.\", it.val)\n \t\t\t} else {\n \t\t\t\tp.bug(\"Expected float value, but got '%s'.\", it.val)\n@@ -252,7 +252,7 @@ func (p *parser) establishContext(key Key, array bool) {\n \t\tcase map[string]interface{}:\n \t\t\thashContext = t\n \t\tdefault:\n-\t\t\tp.panic(\"Key '%s' was already created as a hash.\", keyContext)\n+\t\t\tp.panicf(\"Key '%s' was already created as a hash.\", keyContext)\n \t\t}\n \t}\n \n@@ -270,7 +270,7 @@ func (p *parser) establishContext(key Key, array bool) {\n \t\tif hash, ok := hashContext[k].([]map[string]interface{}); ok {\n \t\t\thashContext[k] = append(hash, make(map[string]interface{}))\n \t\t} else {\n-\t\t\tp.panic(\"Key '%s' was already created and cannot be used as \"+\n+\t\t\tp.panicf(\"Key '%s' was already created and cannot be used as \"+\n \t\t\t\t\"an array.\", keyContext)\n \t\t}\n \t} else {\n@@ -326,7 +326,7 @@ func (p *parser) setValue(key string, value interface{}) {\n \n \t\t// Otherwise, we have a concrete key trying to override a previous\n \t\t// key, which is *always* wrong.\n-\t\tp.panic(\"Key '%s' has already been defined.\", keyContext)\n+\t\tp.panicf(\"Key '%s' has already been defined.\", keyContext)\n \t}\n \thash[key] = value\n }\n@@ -411,7 +411,7 @@ func (p *parser) asciiEscapeToUnicode(s string) string {\n \t// UTF-8 characters like U+DCFF, but it doesn't.\n \tr := string(rune(hex))\n \tif !utf8.ValidString(r) {\n-\t\tp.panic(\"Escaped character '\\\\u%s' is not valid UTF-8.\", s)\n+\t\tp.panicf(\"Escaped character '\\\\u%s' is not valid UTF-8.\", s)\n \t}\n \treturn string(r)\n }")),
			}),
			(github.CommitFile)(github.CommitFile{
				SHA:       (*string)(NewString("b7897e79d2b19c3690d9e34e0c5fe7e71b5fd680")),
				Filename:  (*string)(NewString("toml-test-encoder/main.go")),
				Additions: (*int)(NewInt(2)),
				Deletions: (*int)(NewInt(2)),
				Changes:   (*int)(NewInt(4)),
				Status:    (*string)(NewString("modified")),
				Patch:     (*string)(NewString("@@ -59,8 +59,8 @@ func translate(typedJson interface{}) interface{} {\n \t\t\tif m, ok := translate(v[i]).(map[string]interface{}); ok {\n \t\t\t\ttabArray[i] = m\n \t\t\t} else {\n-\t\t\t\tlog.Fatalf(\"JSON arrays may only contain objects. This \"+\n-\t\t\t\t\t\"corresponds to only tables being allowed in \"+\n+\t\t\t\tlog.Fatalf(\"JSON arrays may only contain objects. This \" +\n+\t\t\t\t\t\"corresponds to only tables being allowed in \" +\n \t\t\t\t\t\"TOML table arrays.\")\n \t\t\t}\n \t\t}")),
			}),
			(github.CommitFile)(github.CommitFile{
				SHA:       (*string)(NewString("026ac6ae6d52914510b2a647f85ea513c3e822d9")),
				Filename:  (*string)(NewString("type_check.go")),
				Additions: (*int)(NewInt(1)),
				Deletions: (*int)(NewInt(1)),
				Changes:   (*int)(NewInt(2)),
				Status:    (*string)(NewString("modified")),
				Patch:     (*string)(NewString("@@ -70,7 +70,7 @@ func (p *parser) typeOfArray(types []tomlType) tomlType {\n \ttheType := types[0]\n \tfor _, t := range types[1:] {\n \t\tif !typeEqual(theType, t) {\n-\t\t\tp.panic(\"Array contains values of type '%s' and '%s', but arrays \"+\n+\t\t\tp.panicf(\"Array contains values of type '%s' and '%s', but arrays \"+\n \t\t\t\t\"must be homogeneous.\", theType, t)\n \t\t}\n \t}")),
			}),
		}),
	})

	GenerateGithubHtml(buf, goPackage, cc)

	CheckError(http.ListenAndServe(":8081", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.Copy(io.MultiWriter(w, os.Stdout), bytes.NewReader(buf.Bytes()))
	})))

	// Output:
	//<h3>github.com/BurntSushi/toml</h3><form name="x-update" method="POST" action="/-/update"><input type="hidden" name="import_path" value="github.com/BurntSushi/toml"></form><a href="javascript:document.getElementsByName('x-update')[0].submit();" title="go get -u -d github.com/BurntSushi/toml">Update</a><ol><li>We want %s since errorf escapes some characters (like new lines), which turns them into strings.</li><li>fix go vet warnings</li><li>gofmt</li></ol>
}
