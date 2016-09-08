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
	got, err := ParseReader("stdin", os.Stdin, Debug(true))
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

func toMatchRef(m interface{}) *tmatch {
	if m == nil {
		return nil
	}
	r := m.(tmatch)
	fmt.Println(r.pos.pos)
	return &r
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Trace",
			pos:  position{line: 73, col: 1, offset: 1140},
			expr: &seqExpr{
				pos: position{line: 73, col: 10, offset: 1149},
				exprs: []interface{}{
					&stateCodeExpr{
						pos: position{line: 73, col: 10, offset: 1149},
						run: (*parser).callonTrace2,
					},
					&zeroOrMoreExpr{
						pos: position{line: 73, col: 43, offset: 1182},
						expr: &ruleRefExpr{
							pos:  position{line: 73, col: 44, offset: 1183},
							name: "TraceLine",
						},
					},
				},
			},
		},
		{
			name: "TraceLine",
			pos:  position{line: 74, col: 1, offset: 1196},
			expr: &seqExpr{
				pos: position{line: 74, col: 14, offset: 1209},
				exprs: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 74, col: 14, offset: 1209},
						name: "INDENTATION",
					},
					&ruleRefExpr{
						pos:  position{line: 74, col: 26, offset: 1221},
						name: "TraceEntry",
					},
				},
			},
		},
		{
			name: "TraceEntry",
			pos:  position{line: 75, col: 1, offset: 1235},
			expr: &actionExpr{
				pos: position{line: 75, col: 16, offset: 1250},
				run: (*parser).callonTraceEntry1,
				expr: &seqExpr{
					pos: position{line: 75, col: 16, offset: 1250},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 75, col: 16, offset: 1250},
							name: "INDENT",
						},
						&labeledExpr{
							pos:   position{line: 75, col: 23, offset: 1257},
							label: "enter",
							expr: &ruleRefExpr{
								pos:  position{line: 75, col: 29, offset: 1263},
								name: "EnterParseEntry",
							},
						},
						&labeledExpr{
							pos:   position{line: 75, col: 45, offset: 1279},
							label: "calls",
							expr: &zeroOrMoreExpr{
								pos: position{line: 75, col: 51, offset: 1285},
								expr: &choiceExpr{
									pos: position{line: 75, col: 52, offset: 1286},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 75, col: 52, offset: 1286},
											name: "TraceEntry",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 65, offset: 1299},
											name: "RestoreEntry",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 75, col: 80, offset: 1314},
							label: "match",
							expr: &zeroOrOneExpr{
								pos: position{line: 75, col: 86, offset: 1320},
								expr: &ruleRefExpr{
									pos:  position{line: 75, col: 86, offset: 1320},
									name: "MatchEntry",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 75, col: 98, offset: 1332},
							name: "DEDENT",
						},
						&ruleRefExpr{
							pos:  position{line: 75, col: 105, offset: 1339},
							name: "ExitParseEntry",
						},
					},
				},
			},
		},
		{
			name: "EnterParseEntry",
			pos:  position{line: 77, col: 1, offset: 1446},
			expr: &actionExpr{
				pos: position{line: 77, col: 20, offset: 1465},
				run: (*parser).callonEnterParseEntry1,
				expr: &seqExpr{
					pos: position{line: 77, col: 20, offset: 1465},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 77, col: 20, offset: 1465},
							expr: &ruleRefExpr{
								pos:  position{line: 77, col: 20, offset: 1465},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 77, col: 23, offset: 1468},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 77, col: 27, offset: 1472},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 77, col: 29, offset: 1474},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 77, col: 33, offset: 1478},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 77, col: 42, offset: 1487},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 77, col: 44, offset: 1489},
							val:        "parse",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 77, col: 52, offset: 1497},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 57, offset: 1502},
								expr: &seqExpr{
									pos: position{line: 77, col: 58, offset: 1503},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 77, col: 58, offset: 1503},
											expr: &ruleRefExpr{
												pos:  position{line: 77, col: 59, offset: 1504},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 77, col: 66, offset: 1511,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 77, col: 70, offset: 1515},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 77, col: 74, offset: 1519},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 77, col: 81, offset: 1526},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "ExitParseEntry",
			pos:  position{line: 79, col: 1, offset: 1631},
			expr: &actionExpr{
				pos: position{line: 79, col: 20, offset: 1650},
				run: (*parser).callonExitParseEntry1,
				expr: &seqExpr{
					pos: position{line: 79, col: 20, offset: 1650},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 79, col: 20, offset: 1650},
							expr: &ruleRefExpr{
								pos:  position{line: 79, col: 20, offset: 1650},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 79, col: 23, offset: 1653},
							val:        "<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 79, col: 27, offset: 1657},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 79, col: 29, offset: 1659},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 79, col: 33, offset: 1663},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 79, col: 42, offset: 1672},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 79, col: 44, offset: 1674},
							val:        "parse",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 79, col: 52, offset: 1682},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 79, col: 57, offset: 1687},
								expr: &seqExpr{
									pos: position{line: 79, col: 58, offset: 1688},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 79, col: 58, offset: 1688},
											expr: &ruleRefExpr{
												pos:  position{line: 79, col: 59, offset: 1689},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 79, col: 66, offset: 1696,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 79, col: 70, offset: 1700},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 79, col: 74, offset: 1704},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 79, col: 81, offset: 1711},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "RestoreEntry",
			pos:  position{line: 81, col: 1, offset: 1816},
			expr: &actionExpr{
				pos: position{line: 81, col: 17, offset: 1832},
				run: (*parser).callonRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 81, col: 17, offset: 1832},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 81, col: 17, offset: 1832},
							label: "enter",
							expr: &ruleRefExpr{
								pos:  position{line: 81, col: 23, offset: 1838},
								name: "EnterRestoreEntry",
							},
						},
						&labeledExpr{
							pos:   position{line: 81, col: 41, offset: 1856},
							label: "exit",
							expr: &ruleRefExpr{
								pos:  position{line: 81, col: 46, offset: 1861},
								name: "ExitRestoreEntry",
							},
						},
					},
				},
			},
		},
		{
			name: "EnterRestoreEntry",
			pos:  position{line: 82, col: 1, offset: 1940},
			expr: &actionExpr{
				pos: position{line: 82, col: 25, offset: 1964},
				run: (*parser).callonEnterRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 82, col: 25, offset: 1964},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 82, col: 25, offset: 1964},
							expr: &ruleRefExpr{
								pos:  position{line: 82, col: 25, offset: 1964},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 82, col: 28, offset: 1967},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 82, col: 32, offset: 1971},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 82, col: 34, offset: 1973},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 82, col: 38, offset: 1977},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 82, col: 47, offset: 1986},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 82, col: 49, offset: 1988},
							val:        "restore",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 82, col: 59, offset: 1998},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 82, col: 61, offset: 2000},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 82, col: 65, offset: 2004},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 82, col: 72, offset: 2011},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "ExitRestoreEntry",
			pos:  position{line: 84, col: 1, offset: 2079},
			expr: &actionExpr{
				pos: position{line: 84, col: 25, offset: 2103},
				run: (*parser).callonExitRestoreEntry1,
				expr: &seqExpr{
					pos: position{line: 84, col: 25, offset: 2103},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 84, col: 25, offset: 2103},
							expr: &ruleRefExpr{
								pos:  position{line: 84, col: 25, offset: 2103},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 84, col: 28, offset: 2106},
							val:        "<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 84, col: 32, offset: 2110},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 84, col: 34, offset: 2112},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 84, col: 38, offset: 2116},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 84, col: 47, offset: 2125},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 84, col: 49, offset: 2127},
							val:        "restore",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 84, col: 59, offset: 2137},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 84, col: 61, offset: 2139},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 84, col: 65, offset: 2143},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 84, col: 72, offset: 2150},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "MatchEntry",
			pos:  position{line: 86, col: 1, offset: 2218},
			expr: &actionExpr{
				pos: position{line: 86, col: 15, offset: 2232},
				run: (*parser).callonMatchEntry1,
				expr: &seqExpr{
					pos: position{line: 86, col: 15, offset: 2232},
					exprs: []interface{}{
						&zeroOrMoreExpr{
							pos: position{line: 86, col: 15, offset: 2232},
							expr: &ruleRefExpr{
								pos:  position{line: 86, col: 15, offset: 2232},
								name: "_",
							},
						},
						&litMatcher{
							pos:        position{line: 86, col: 18, offset: 2235},
							val:        "MATCH",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 86, col: 26, offset: 2243},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 86, col: 28, offset: 2245},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 86, col: 32, offset: 2249},
								name: "Position",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 86, col: 41, offset: 2258},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 86, col: 43, offset: 2260},
							label: "text",
							expr: &zeroOrMoreExpr{
								pos: position{line: 86, col: 48, offset: 2265},
								expr: &seqExpr{
									pos: position{line: 86, col: 49, offset: 2266},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 86, col: 49, offset: 2266},
											expr: &ruleRefExpr{
												pos:  position{line: 86, col: 50, offset: 2267},
												name: "Cursor",
											},
										},
										&anyMatcher{
											line: 86, col: 57, offset: 2274,
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 86, col: 61, offset: 2278},
							label: "cur",
							expr: &ruleRefExpr{
								pos:  position{line: 86, col: 65, offset: 2282},
								name: "Cursor",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 86, col: 72, offset: 2289},
							name: "NL",
						},
					},
				},
			},
		},
		{
			name: "Cursor",
			pos:  position{line: 89, col: 1, offset: 2391},
			expr: &actionExpr{
				pos: position{line: 89, col: 11, offset: 2401},
				run: (*parser).callonCursor1,
				expr: &seqExpr{
					pos: position{line: 89, col: 11, offset: 2401},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 89, col: 11, offset: 2401},
							val:        "[U+",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 89, col: 17, offset: 2407},
							label: "n",
							expr: &ruleRefExpr{
								pos:  position{line: 89, col: 19, offset: 2409},
								name: "HexNumber",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 89, col: 29, offset: 2419},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 89, col: 31, offset: 2421},
							val:        "'",
							ignoreCase: false,
						},
						&anyMatcher{
							line: 89, col: 35, offset: 2425,
						},
						&litMatcher{
							pos:        position{line: 89, col: 37, offset: 2427},
							val:        "']",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Position",
			pos:  position{line: 90, col: 1, offset: 2451},
			expr: &actionExpr{
				pos: position{line: 90, col: 13, offset: 2463},
				run: (*parser).callonPosition1,
				expr: &seqExpr{
					pos: position{line: 90, col: 13, offset: 2463},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 90, col: 13, offset: 2463},
							label: "line",
							expr: &ruleRefExpr{
								pos:  position{line: 90, col: 18, offset: 2468},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 90, col: 25, offset: 2475},
							val:        ":",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 90, col: 29, offset: 2479},
							label: "col",
							expr: &ruleRefExpr{
								pos:  position{line: 90, col: 33, offset: 2483},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 90, col: 40, offset: 2490},
							val:        ":",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 90, col: 44, offset: 2494},
							label: "pos",
							expr: &ruleRefExpr{
								pos:  position{line: 90, col: 48, offset: 2498},
								name: "Number",
							},
						},
						&litMatcher{
							pos:        position{line: 90, col: 55, offset: 2505},
							val:        ":",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Number",
			pos:  position{line: 91, col: 1, offset: 2568},
			expr: &actionExpr{
				pos: position{line: 91, col: 11, offset: 2578},
				run: (*parser).callonNumber1,
				expr: &oneOrMoreExpr{
					pos: position{line: 91, col: 11, offset: 2578},
					expr: &charClassMatcher{
						pos:        position{line: 91, col: 11, offset: 2578},
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
			pos:  position{line: 92, col: 1, offset: 2658},
			expr: &actionExpr{
				pos: position{line: 92, col: 14, offset: 2671},
				run: (*parser).callonHexNumber1,
				expr: &oneOrMoreExpr{
					pos: position{line: 92, col: 14, offset: 2671},
					expr: &charClassMatcher{
						pos:        position{line: 92, col: 14, offset: 2671},
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
			pos:  position{line: 93, col: 1, offset: 2755},
			expr: &actionExpr{
				pos: position{line: 93, col: 7, offset: 2761},
				run: (*parser).callonNL1,
				expr: &seqExpr{
					pos: position{line: 93, col: 7, offset: 2761},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 93, col: 7, offset: 2761},
							expr: &litMatcher{
								pos:        position{line: 93, col: 7, offset: 2761},
								val:        "\r",
								ignoreCase: false,
							},
						},
						&litMatcher{
							pos:        position{line: 93, col: 13, offset: 2767},
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
			pos:         position{line: 94, col: 1, offset: 2792},
			expr: &litMatcher{
				pos:        position{line: 94, col: 19, offset: 2810},
				val:        " ",
				ignoreCase: false,
			},
		},
		{
			name: "INDENTATION",
			pos:  position{line: 97, col: 1, offset: 2819},
			expr: &seqExpr{
				pos: position{line: 97, col: 15, offset: 2835},
				exprs: []interface{}{
					&labeledExpr{
						pos:   position{line: 97, col: 15, offset: 2835},
						label: "spaces",
						expr: &zeroOrMoreExpr{
							pos: position{line: 97, col: 22, offset: 2842},
							expr: &litMatcher{
								pos:        position{line: 97, col: 22, offset: 2842},
								val:        " ",
								ignoreCase: false,
							},
						},
					},
					&andCodeExpr{
						pos: position{line: 97, col: 27, offset: 2847},
						run: (*parser).callonINDENTATION5,
					},
				},
			},
		},
		{
			name: "INDENT",
			pos:  position{line: 98, col: 1, offset: 2971},
			expr: &stateCodeExpr{
				pos: position{line: 98, col: 11, offset: 2981},
				run: (*parser).callonINDENT1,
			},
		},
		{
			name: "DEDENT",
			pos:  position{line: 99, col: 1, offset: 3037},
			expr: &stateCodeExpr{
				pos: position{line: 99, col: 11, offset: 3047},
				run: (*parser).callonDEDENT1,
			},
		},
	},
}

/*c*/
func (state statedict) onTrace2() error {
	state["indent"] = 1
	return nil
}

func (p *parser) callonTrace2() (bool, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	err := p.pt.state.onTrace2()
	copyState(state, p.pt.state)
	return true, err
}

func (c *current) onTraceEntry1(enter, calls, match interface{}) (interface{}, error) {
	return traceentry{enter.(tparse), calls.([]interface{}), toMatchRef(match)}, nil
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

func (c *current) onINDENTATION5(spaces interface{}) (bool, error) {
	fmt.Println(len(toIfaceSlice(spaces)), state["indent"])
	return len(toIfaceSlice(spaces)) == state["indent"].(int), nil
}

func (p *parser) callonINDENTATION5() (bool, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onINDENTATION5(stack["spaces"])
}

/*c*/
func (state statedict) onINDENT1() error {
	state["indent"] = state["indent"].(int) + 1
	return nil
}

func (p *parser) callonINDENT1() (bool, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	err := p.pt.state.onINDENT1()
	copyState(state, p.pt.state)
	return true, err
}

/*c*/
func (state statedict) onDEDENT1() error {
	state["indent"] = state["indent"].(int) - 1
	return nil
}

func (p *parser) callonDEDENT1() (bool, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	err := p.pt.state.onDEDENT1()
	copyState(state, p.pt.state)
	return true, err
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")

	// State
	state = make(statedict)
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
	rn    rune
	w     int
	state statedict
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

type statedict map[string]interface{}

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

type stateCodeExpr struct {
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
		pt:       savepoint{position: position{line: 1}, state: make(statedict)},
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
		if n == 1 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// copy state
func copyState(dst, src statedict) {
	for k, v := range src {
		dst[k] = v
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
	case *stateCodeExpr:
		val, ok = p.parseStateCodeExpr(expr)
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
	pt.state = make(statedict)
	copyState(pt.state, p.pt.state)
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	copyState(p.pt.state, pt.state)
	copyState(state, pt.state)
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

func (p *parser) parseStateCodeExpr(state *stateCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseStateCodeExpr"))
	}

	_, err := state.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, true
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	pt.state = make(statedict)
	copyState(pt.state, p.pt.state)
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	copyState(p.pt.state, pt.state)
	copyState(state, pt.state)
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
	pt.state = make(statedict)
	copyState(pt.state, p.pt.state)
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			copyState(p.pt.state, pt.state)
			copyState(state, pt.state)
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
