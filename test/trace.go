package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	got, err := ParseReader("stdin", os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
func IfaceSliceToString(v []interface{}) string {
	if v == nil {
		return ""
	}
	buffer := bytes.Buffer{}
	for i := range v {
		buffer.WriteString(string(toIfaceSlice(v[i])[1].([]uint8)))
	}
	return buffer.String()
}

type tposition struct {
	line, col, pos int
}
type tcursor struct {
	codepoint int
}

type tparse struct {
	pos    tposition
	text   string
	cursor int
	dir    string
}

type tmatch struct {
	pos    tposition
	text   string
	cursor int
}

type trestore struct {
	pos    tposition
	cursor int
	dir    string
}

type traceentry struct {
	value tparse
	calls []interface{}
	match *tmatch
}

func toMatch(m interface{}) *tmatch {
	if m == nil {
		return nil
	}
	r := m.(tmatch)
	return &r
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Trace",
			pos:  position{line: 72, col: 1, offset: 1096},
			expr: &zeroOrMoreExpr{
				pos: position{line: 72, col: 10, offset: 1105},
				expr: &ruleRefExpr{
					pos:  position{line: 72, col: 11, offset: 1106},
					name: "TraceLine",
				},
			},
		},
		{
			name: "TraceLine",
			pos:  position{line: 73, col: 1, offset: 1119},
			expr: &ruleRefExpr{
				pos:  position{line: 73, col: 14, offset: 1132},
				name: "TraceEntry",
			},
		},
		{
			name: "TraceEntry",
			pos:  position{line: 74, col: 1, offset: 1144},
			expr: &actionExpr{
				pos: position{line: 74, col: 15, offset: 1158},
				run: (*parser).callonTraceEntry1,
				expr: &seqExpr{
					pos: position{line: 74, col: 15, offset: 1158},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 74, col: 15, offset: 1158},
							label: "enter",
							expr: &ruleRefExpr{
								pos:  position{line: 74, col: 21, offset: 1164},
								name: "EnterParseEntry",
							},
						},
						&labeledExpr{
							pos:   position{line: 74, col: 37, offset: 1180},
							label: "calls",
							expr: &zeroOrMoreExpr{
								pos: position{line: 74, col: 43, offset: 1186},
								expr: &choiceExpr{
									pos: position{line: 74, col: 44, offset: 1187},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 74, col: 44, offset: 1187},
											name: "TraceEntry",
										},
										&ruleRefExpr{
											pos:  position{line: 74, col: 57, offset: 1200},
											name: "RestoreEntry",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 74, col: 72, offset: 1215},
							label: "match",
							expr: &zeroOrOneExpr{
								pos: position{line: 74, col: 78, offset: 1221},
								expr: &ruleRefExpr{
									pos:  position{line: 74, col: 78, offset: 1221},
									name: "MatchEntry",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 74, col: 90, offset: 1233},
							name: "ExitParseEntry",
						},
					},
				},
			},
		},
		{
			name: "EnterParseEntry",
			pos:  position{line: 76, col: 1, offset: 1337},
			expr: &actionExpr{
				pos: position{line: 76, col: 20, offset: 1356},
				run: (*parser).callonEnterParseEntry1,
				expr: &seqExpr{
					pos: position{line: 76, col: 20, offset: 1356},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 76, col: 20, offset: 1356},
							expr: &ruleRefExpr{
								pos:  position{line: 76, col: 20, offset: 1356},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 76, col: 23, offset: 1359},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 76, col: 27, offset: 1363},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 76, col: 29, offset: 1365},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 76, col: 33, offset: 1369},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 76, col: 42, offset: 1378},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 76, col: 44, offset: 1380},
							val:        "parse",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 76, col: 52, offset: 1388},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 76, col: 57, offset: 1393},
								expr: &seqExpr{
									pos: position{line: 76, col: 58, offset: 1394},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 76, col: 58, offset: 1394},
											expr: &ruleRefExpr{
												pos:  position{line: 76, col: 59, offset: 1395},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 76, col: 66, offset: 1402,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 76, col: 70, offset: 1406},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 76, col: 74, offset: 1410},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 76, col: 81, offset: 1417},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "ExitParseEntry",
			pos:  position{line: 78, col: 1, offset: 1522},
			expr: &actionExpr{
				pos: position{line: 78, col: 20, offset: 1541},
				run: (*parser).callonExitParseEntry1,
				expr: &seqExpr{
					pos: position{line: 78, col: 20, offset: 1541},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 78, col: 20, offset: 1541},
							expr: &ruleRefExpr{
								pos:  position{line: 78, col: 20, offset: 1541},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 78, col: 23, offset: 1544},
							val:        "<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 27, offset: 1548},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 78, col: 29, offset: 1550},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 78, col: 33, offset: 1554},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 42, offset: 1563},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 78, col: 44, offset: 1565},
							val:        "parse",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 78, col: 52, offset: 1573},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 78, col: 57, offset: 1578},
								expr: &seqExpr{
									pos: position{line: 78, col: 58, offset: 1579},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 78, col: 58, offset: 1579},
											expr: &ruleRefExpr{
												pos:  position{line: 78, col: 59, offset: 1580},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 78, col: 66, offset: 1587,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 78, col: 70, offset: 1591},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 78, col: 74, offset: 1595},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 81, offset: 1602},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "RestoreEntry",
			pos:  position{line: 80, col: 1, offset: 1707},
			expr: &actionExpr{
				pos: position{line: 80, col: 17, offset: 1723},
				run: (*parser).callonRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 80, col: 17, offset: 1723},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 80, col: 17, offset: 1723},
							label: "enter",
							expr: &ruleRefExpr{
								pos:  position{line: 80, col: 23, offset: 1729},
								name: "EnterRestoreEntry",
							},
						},
						&labeledExpr{
							pos:   position{line: 80, col: 41, offset: 1747},
							label: "exit",
							expr: &ruleRefExpr{
								pos:  position{line: 80, col: 46, offset: 1752},
								name: "ExitRestoreEntry",
							},
						},
					},
				},
			},
		},
		{
			name: "EnterRestoreEntry",
			pos:  position{line: 81, col: 1, offset: 1831},
			expr: &actionExpr{
				pos: position{line: 81, col: 25, offset: 1855},
				run: (*parser).callonEnterRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 81, col: 25, offset: 1855},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 81, col: 25, offset: 1855},
							expr: &ruleRefExpr{
								pos:  position{line: 81, col: 25, offset: 1855},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 81, col: 28, offset: 1858},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 81, col: 32, offset: 1862},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 81, col: 34, offset: 1864},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 81, col: 38, offset: 1868},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 81, col: 47, offset: 1877},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 81, col: 49, offset: 1879},
							val:        "restore",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 81, col: 59, offset: 1889},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 81, col: 61, offset: 1891},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 81, col: 65, offset: 1895},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 81, col: 72, offset: 1902},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "ExitRestoreEntry",
			pos:  position{line: 83, col: 1, offset: 1970},
			expr: &actionExpr{
				pos: position{line: 83, col: 25, offset: 1994},
				run: (*parser).callonExitRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 83, col: 25, offset: 1994},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 83, col: 25, offset: 1994},
							expr: &ruleRefExpr{
								pos:  position{line: 83, col: 25, offset: 1994},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 83, col: 28, offset: 1997},
							val:        "<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 83, col: 32, offset: 2001},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 83, col: 34, offset: 2003},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 83, col: 38, offset: 2007},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 83, col: 47, offset: 2016},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 83, col: 49, offset: 2018},
							val:        "restore",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 83, col: 59, offset: 2028},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 83, col: 61, offset: 2030},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 83, col: 65, offset: 2034},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 83, col: 72, offset: 2041},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "MatchEntry",
			pos:  position{line: 85, col: 1, offset: 2109},
			expr: &actionExpr{
				pos: position{line: 85, col: 15, offset: 2123},
				run: (*parser).callonMatchEntry1,
				expr: &seqExpr{
					pos: position{line: 85, col: 15, offset: 2123},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 85, col: 15, offset: 2123},
							expr: &ruleRefExpr{
								pos:  position{line: 85, col: 15, offset: 2123},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 85, col: 18, offset: 2126},
							val:        "MATCH",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 26, offset: 2134},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 28, offset: 2136},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 85, col: 32, offset: 2140},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 41, offset: 2149},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 43, offset: 2151},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 48, offset: 2156},
								expr: &seqExpr{
									pos: position{line: 85, col: 49, offset: 2157},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 85, col: 49, offset: 2157},
											expr: &ruleRefExpr{
												pos:  position{line: 85, col: 50, offset: 2158},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 85, col: 57, offset: 2165,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 85, col: 61, offset: 2169},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 85, col: 65, offset: 2173},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 72, offset: 2180},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "Cursor",
			pos:  position{line: 88, col: 1, offset: 2282},
			expr: &actionExpr{
				pos: position{line: 88, col: 11, offset: 2292},
				run: (*parser).callonCursor1,
				expr: &seqExpr{
					pos: position{line: 88, col: 11, offset: 2292},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 88, col: 11, offset: 2292},
							val:        "[U+",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 88, col: 17, offset: 2298},
							label: "n",
							expr: &ruleRefExpr{
								pos:  position{line: 88, col: 19, offset: 2300},
								name: "HexNumber",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 88, col: 29, offset: 2310},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 88, col: 31, offset: 2312},
							val:        "'",
							ignoreCase: false,
						},
						&anyMatcher{
							line: 88, col: 35, offset: 2316,
						},
						&litMatcher{
							pos:        position{line: 88, col: 37, offset: 2318},
							val:        "']",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Position",
			pos:  position{line: 89, col: 1, offset: 2342},
			expr: &actionExpr{
				pos: position{line: 89, col: 13, offset: 2354},
				run: (*parser).callonPosition1,
				expr: &seqExpr{
					pos: position{line: 89, col: 13, offset: 2354},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 89, col: 13, offset: 2354},
							label: "line",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 18, offset: 2359},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 89, col: 25, offset: 2366},
							val:        ":",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 89, col: 29, offset: 2370},
							label: "col",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 33, offset: 2374},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 89, col: 40, offset: 2381},
							val:        ":",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 89, col: 44, offset: 2385},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 48, offset: 2389},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 89, col: 55, offset: 2396},
							val:        ":",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 90, col: 1, offset: 2459},
			expr: &actionExpr{
				pos: position{line: 90, col: 11, offset: 2469},
				run: (*parser).callonNumber1,
				expr: &oneOrMoreExpr{
					pos: position{line: 90, col: 11, offset: 2469},
					expr: &charClassMatcher{
						pos:        position{line: 90, col: 11, offset: 2469},
						val:        "[0-9]",
						ranges:     []rune{'0', '9'},
						ignoreCase: false,
						inverted:   false,
					},
				},
			},
		},
		{
			name: "HexNumber",
			pos:  position{line: 91, col: 1, offset: 2549},
			expr: &actionExpr{
				pos: position{line: 91, col: 14, offset: 2562},
				run: (*parser).callonHexNumber1,
				expr: &oneOrMoreExpr{
					pos: position{line: 91, col: 14, offset: 2562},
					expr: &charClassMatcher{
						pos:        position{line: 91, col: 14, offset: 2562},
						val:        "[0-9A-F]",
						ranges:     []rune{'0', '9', 'A', 'F'},
						ignoreCase: false,
						inverted:   false,
					},
				},
			},
		},
		{
			name: "NL",
			pos:  position{line: 92, col: 1, offset: 2646},
			expr: &actionExpr{
				pos: position{line: 92, col: 7, offset: 2652},
				run: (*parser).callonNL1,
				expr: &seqExpr{
					pos: position{line: 92, col: 7, offset: 2652},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 92, col: 7, offset: 2652},
							expr: &litMatcher{
								pos:        position{line: 92, col: 7, offset: 2652},
								val:        "\r",
								ignoreCase: false,
							},
						},
						&litMatcher{
							pos:        position{line: 92, col: 13, offset: 2658},
							val:        "\n",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name:        "_",
			displayName: "\"whitespace\"",
			pos:         position{line: 93, col: 1, offset: 2683},
			expr: &litMatcher{
				pos:        position{line: 93, col: 19, offset: 2701},
				val:        " ",
				ignoreCase: false,
			},
		},
	},
}

func (c *current) onTraceEntry1(enter, calls, match interface{}) (interface{}, error) {
	return traceentry{enter.(tparse), calls.([]interface{}), toMatch(match)}, nil
}

func (p *parser) callonTraceEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTraceEntry1(stack["enter"], stack["calls"], stack["match"])
}

func (c *current) onEnterParseEntry1(pos, text, cur interface{}) (interface{}, error) {
	return tparse{pos.(tposition), IfaceSliceToString(toIfaceSlice(text)), cur.(int), ">"}, nil
}

func (p *parser) callonEnterParseEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnterParseEntry1(stack["pos"], stack["text"], stack["cur"])
}

func (c *current) onExitParseEntry1(pos, text, cur interface{}) (interface{}, error) {
	return tparse{pos.(tposition), IfaceSliceToString(toIfaceSlice(text)), cur.(int), "<"}, nil
}

func (p *parser) callonExitParseEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onExitParseEntry1(stack["pos"], stack["text"], stack["cur"])
}

func (c *current) onRestoreEntry1(enter, exit interface{}) (interface{}, error) {
	return []trestore{enter.(trestore), exit.(trestore)}, nil
}

func (p *parser) callonRestoreEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onRestoreEntry1(stack["enter"], stack["exit"])
}

func (c *current) onEnterRestoreEntry1(pos, cur interface{}) (interface{}, error) {
	return trestore{pos.(tposition), cur.(int), ">"}, nil
}

func (p *parser) callonEnterRestoreEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnterRestoreEntry1(stack["pos"], stack["cur"])
}

func (c *current) onExitRestoreEntry1(pos, cur interface{}) (interface{}, error) {
	return trestore{pos.(tposition), cur.(int), "<"}, nil
}

func (p *parser) callonExitRestoreEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onExitRestoreEntry1(stack["pos"], stack["cur"])
}

func (c *current) onMatchEntry1(pos, text, cur interface{}) (interface{}, error) {
	return tmatch{pos.(tposition), IfaceSliceToString(toIfaceSlice(text)), cur.(int)}, nil
}

func (p *parser) callonMatchEntry1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMatchEntry1(stack["pos"], stack["text"], stack["cur"])
}

func (c *current) onCursor1(n interface{}) (interface{}, error) {
	return n, nil
}

func (p *parser) callonCursor1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCursor1(stack["n"])
}

func (c *current) onPosition1(line, col, pos interface{}) (interface{}, error) {
	return tposition{line.(int), col.(int), pos.(int)}, nil
}

func (p *parser) callonPosition1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPosition1(stack["line"], stack["col"], stack["pos"])
}

func (c *current) onNumber1() (interface{}, error) {
	n, err := strconv.ParseInt(string(c.text), 0, 64)
	return int(n), err
}

func (p *parser) callonNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNumber1()
}

func (c *current) onHexNumber1() (interface{}, error) {
	n, err := strconv.ParseInt(string(c.text), 16, 64)
	return int(n), err
}

func (p *parser) callonHexNumber1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onHexNumber1()
}

func (c *current) onNL1() (interface{}, error) {
	return "", nil
}

func (p *parser) callonNL1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNL1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug   bool
	depth   int

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
	p.errs.add(pe)
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n > 0 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
