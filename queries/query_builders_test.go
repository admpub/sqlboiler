package queries

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/admpub/sqlboiler/drivers"
)

var writeGoldenFiles = flag.Bool(
	"test.golden",
	false,
	"Write golden files.",
)

func TestBuildQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q    *Query
		args []interface{}
	}{
		{&Query{from: []string{"t"}}, nil},
		{&Query{from: []string{"q"}, limit: 5, offset: 6}, nil},
		{&Query{from: []string{"q"}, orderBy: []string{"a ASC", "b DESC"}}, nil},
		{&Query{from: []string{"t"}, selectCols: []string{"count(*) as ab, thing as bd", `"stuff"`}}, nil},
		{&Query{from: []string{"a", "b"}, selectCols: []string{"count(*) as ab, thing as bd", `"stuff"`}}, nil},
		{&Query{
			selectCols: []string{"a.happy", "r.fun", "q"},
			from:       []string{"happiness as a"},
			joins:      []join{{clause: "rainbows r on a.id = r.happy_id"}},
		}, nil},
		{&Query{
			from:  []string{"happiness as a"},
			joins: []join{{clause: "rainbows r on a.id = r.happy_id"}},
		}, nil},
		{&Query{
			from: []string{"videos"},
			joins: []join{{
				clause: "(select id from users where deleted = ?) u on u.id = videos.user_id",
				args:   []interface{}{true},
			}},
			where: []where{{clause: "videos.deleted = ?", args: []interface{}{false}}},
		}, []interface{}{true, false}},
		{&Query{
			from:    []string{"a"},
			groupBy: []string{"id", "name"},
			where: []where{
				{clause: "a=? or b=?", args: []interface{}{1, 2}},
				{clause: "c=?", args: []interface{}{3}},
			},
			having: []having{
				{clause: "id <> ?", args: []interface{}{1}},
				{clause: "length(name, ?) > ?", args: []interface{}{"utf8", 5}},
			},
		}, []interface{}{1, 2, 3, 1, "utf8", 5}},
		{&Query{
			delete: true,
			from:   []string{"thing happy", `upset as "sad"`, "fun", "thing as stuff", `"angry" as mad`},
			where: []where{
				{clause: "a=?", args: []interface{}{1}},
				{clause: "b=?", args: []interface{}{2}},
				{clause: "c=?", args: []interface{}{3}},
			},
		}, []interface{}{1, 2, 3}},
		{&Query{
			delete: true,
			from:   []string{"thing happy", `upset as "sad"`, "fun", "thing as stuff", `"angry" as mad`},
			where: []where{
				{clause: "(id=? and thing=?) or stuff=?", args: []interface{}{1, 2, 3}},
			},
			limit: 5,
		}, []interface{}{1, 2, 3}},
		{&Query{
			from: []string{"thing happy", `"fun"`, `stuff`},
			update: map[string]interface{}{
				"col1":       1,
				`"col2"`:     2,
				`"fun".col3`: 3,
			},
			where: []where{
				{clause: "aa=? or bb=? or cc=?", orSeparator: true, args: []interface{}{4, 5, 6}},
				{clause: "dd=? or ee=? or ff=? and gg=?", args: []interface{}{7, 8, 9, 10}},
			},
			limit: 5,
		}, []interface{}{2, 3, 1, 4, 5, 6, 7, 8, 9, 10}},
		{&Query{from: []string{"cats"}, joins: []join{{JoinInner, "dogs d on d.cat_id = cats.id", nil}}}, nil},
		{&Query{from: []string{"cats c"}, joins: []join{{JoinInner, "dogs d on d.cat_id = cats.id", nil}}}, nil},
		{&Query{from: []string{"cats as c"}, joins: []join{{JoinInner, "dogs d on d.cat_id = cats.id", nil}}}, nil},
		{&Query{from: []string{"cats as c", "dogs as d"}, joins: []join{{JoinInner, "dogs d on d.cat_id = cats.id", nil}}}, nil},
	}

	for i, test := range tests {
		filename := filepath.Join("_fixtures", fmt.Sprintf("%02d.sql", i))
		test.q.dialect = &drivers.Dialect{LQ: '"', RQ: '"', UseIndexPlaceholders: true}
		out, args := BuildQuery(test.q)

		if *writeGoldenFiles {
			err := ioutil.WriteFile(filename, []byte(out), 0664)
			if err != nil {
				t.Fatalf("Failed to write golden file %s: %s\n", filename, err)
			}
			t.Logf("wrote golden file: %s\n", filename)
			continue
		}

		byt, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read golden file %q: %v", filename, err)
		}

		if string(bytes.TrimSpace(byt)) != out {
			t.Errorf("[%02d] Test failed:\nWant:\n%s\nGot:\n%s", i, byt, out)
		}

		if !reflect.DeepEqual(args, test.args) {
			t.Errorf("[%02d] Test failed:\nWant:\n%s\nGot:\n%s", i, spew.Sdump(test.args), spew.Sdump(args))
		}
	}
}

func TestWriteStars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		In  Query
		Out []string
	}{
		{
			In:  Query{from: []string{`a`}},
			Out: []string{`"a".*`},
		},
		{
			In:  Query{from: []string{`a as b`}},
			Out: []string{`"b".*`},
		},
		{
			In:  Query{from: []string{`a as b`, `c`}},
			Out: []string{`"b".*`, `"c".*`},
		},
		{
			In:  Query{from: []string{`a as b`, `c as d`}},
			Out: []string{`"b".*`, `"d".*`},
		},
	}

	for i, test := range tests {
		test.In.dialect = &drivers.Dialect{LQ: '"', RQ: '"', UseIndexPlaceholders: true}
		selects := writeStars(&test.In)
		if !reflect.DeepEqual(selects, test.Out) {
			t.Errorf("writeStar test fail %d\nwant: %v\ngot:  %v", i, test.Out, selects)
		}
	}
}

func TestWhereClause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q      Query
		expect string
	}{
		// Or("a=?")
		{
			q: Query{
				where: []where{{clause: "a=?", orSeparator: true}},
			},
			expect: " WHERE (a=$1)",
		},
		// Where("a=?")
		{
			q: Query{
				where: []where{{clause: "a=?"}},
			},
			expect: " WHERE (a=$1)",
		},
		// Where("(a=?)")
		{
			q: Query{
				where: []where{{clause: "(a=?)"}},
			},
			expect: " WHERE ((a=$1))",
		},
		// Where("((a=? OR b=?))")
		{
			q: Query{
				where: []where{{clause: "((a=? OR b=?))"}},
			},
			expect: " WHERE (((a=$1 OR b=$2)))",
		},
		// Where("(a=?)", Or("(b=?)")
		{
			q: Query{
				where: []where{
					{clause: "(a=?)"},
					{clause: "(b=?)", orSeparator: true},
				},
			},
			expect: " WHERE ((a=$1)) OR ((b=$2))",
		},
		// Where("a=? OR b=?")
		{
			q: Query{
				where: []where{{clause: "a=? OR b=?"}},
			},
			expect: " WHERE (a=$1 OR b=$2)",
		},
		// Where("a=?"), Where("b=?")
		{
			q: Query{
				where: []where{{clause: "a=?"}, {clause: "b=?"}},
			},
			expect: " WHERE (a=$1) AND (b=$2)",
		},
		// Where("(a=? AND b=?) OR c=?")
		{
			q: Query{
				where: []where{{clause: "(a=? AND b=?) OR c=?"}},
			},
			expect: " WHERE ((a=$1 AND b=$2) OR c=$3)",
		},
		// Where("a=? OR b=?"), Where("c=? OR d=? OR e=?")
		{
			q: Query{
				where: []where{
					{clause: "(a=? OR b=?)"},
					{clause: "(c=? OR d=? OR e=?)"},
				},
			},
			expect: " WHERE ((a=$1 OR b=$2)) AND ((c=$3 OR d=$4 OR e=$5))",
		},
		// Where("(a=? AND b=?) OR (c=? AND d=? AND e=?) OR f=? OR f=?")
		{
			q: Query{
				where: []where{
					{clause: "(a=? AND b=?) OR (c=? AND d=? AND e=?) OR f=? OR g=?"},
				},
			},
			expect: " WHERE ((a=$1 AND b=$2) OR (c=$3 AND d=$4 AND e=$5) OR f=$6 OR g=$7)",
		},
		// Where("(a=? AND b=?) OR (c=? AND d=? OR e=?) OR f=? OR g=?")
		{
			q: Query{
				where: []where{
					{clause: "(a=? AND b=?) OR (c=? AND d=? OR e=?) OR f=? OR g=?"},
				},
			},
			expect: " WHERE ((a=$1 AND b=$2) OR (c=$3 AND d=$4 OR e=$5) OR f=$6 OR g=$7)",
		},
		// Where("a=? or b=?"), Or("c=? and d=?"), Or("e=? or f=?")
		{
			q: Query{
				where: []where{
					{clause: "a=? or b=?", orSeparator: true},
					{clause: "c=? and d=?", orSeparator: true},
					{clause: "e=? or f=?", orSeparator: true},
				},
			},
			expect: " WHERE (a=$1 or b=$2) OR (c=$3 and d=$4) OR (e=$5 or f=$6)",
		},
		// Where("a=? or b=?"), Or("c=? and d=?"), Or("e=? or f=?")
		{
			q: Query{
				where: []where{
					{clause: "a=? or b=?"},
					{clause: "c=? and d=?", orSeparator: true},
					{clause: "e=? or f=?"},
				},
			},
			expect: " WHERE (a=$1 or b=$2) OR (c=$3 and d=$4) AND (e=$5 or f=$6)",
		},
	}

	for i, test := range tests {
		test.q.dialect = &drivers.Dialect{LQ: '"', RQ: '"', UseIndexPlaceholders: true}
		result, _ := whereClause(&test.q, 1)
		if result != test.expect {
			t.Errorf("%d) Mismatch between expect and result:\n%s\n%s\n", i, test.expect, result)
		}
	}
}

func TestInClause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		q      Query
		expect string
		args   []interface{}
	}{
		{
			q: Query{
				in: []in{{clause: "a in ?", args: []interface{}{}, orSeparator: true}},
			},
			expect: ` WHERE "a" IN ()`,
		},
		{
			q: Query{
				in: []in{{clause: "a in ?", args: []interface{}{1}, orSeparator: true}},
			},
			expect: ` WHERE "a" IN ($1)`,
			args:   []interface{}{1},
		},
		{
			q: Query{
				in: []in{{clause: "a in ?", args: []interface{}{1, 2, 3}}},
			},
			expect: ` WHERE "a" IN ($1,$2,$3)`,
			args:   []interface{}{1, 2, 3},
		},
		{
			q: Query{
				in: []in{{clause: "? in ?", args: []interface{}{1, 2, 3}}},
			},
			expect: " WHERE $1 IN ($2,$3)",
			args:   []interface{}{1, 2, 3},
		},
		{
			q: Query{
				in: []in{{clause: "( ? , ? ) in ( ? )", orSeparator: true, args: []interface{}{"a", "b", 1, 2, 3, 4}}},
			},
			expect: " WHERE ( $1 , $2 ) IN ( (($3,$4),($5,$6)) )",
			args:   []interface{}{"a", "b", 1, 2, 3, 4},
		},
		{
			q: Query{
				in: []in{{clause: `("a")in(?)`, orSeparator: true, args: []interface{}{1, 2, 3}}},
			},
			expect: ` WHERE ("a") IN (($1,$2,$3))`,
			args:   []interface{}{1, 2, 3},
		},
		{
			q: Query{
				in: []in{{clause: `("a")in?`, args: []interface{}{1}}},
			},
			expect: ` WHERE ("a") IN ($1)`,
			args:   []interface{}{1},
		},
		{
			q: Query{
				where: []where{
					{clause: "a=?", args: []interface{}{1}},
				},
				in: []in{
					{clause: `?,?,"name" in ?`, orSeparator: true, args: []interface{}{"c", "d", 3, 4, 5, 6, 7, 8}},
					{clause: `?,?,"name" in ?`, orSeparator: true, args: []interface{}{"e", "f", 9, 10, 11, 12, 13, 14}},
				},
			},
			expect: ` OR $1,$2,"name" IN (($3,$4,$5),($6,$7,$8)) OR $9,$10,"name" IN (($11,$12,$13),($14,$15,$16))`,
			args:   []interface{}{"c", "d", 3, 4, 5, 6, 7, 8, "e", "f", 9, 10, 11, 12, 13, 14},
		},
		{
			q: Query{
				in: []in{
					{clause: `("a")in`, args: []interface{}{1}},
					{clause: `("a")in?`, orSeparator: true, args: []interface{}{1}},
				},
			},
			expect: ` WHERE ("a")in OR ("a") IN ($1)`,
			args:   []interface{}{1, 1},
		},
		{
			q: Query{
				in: []in{
					{clause: `\?,\? in \?`, args: []interface{}{1}},
					{clause: `\?,\?in \?`, orSeparator: true, args: []interface{}{1}},
				},
			},
			expect: ` WHERE ?,? IN ? OR ?,? IN ?`,
			args:   []interface{}{1, 1},
		},
		{
			q: Query{
				in: []in{
					{clause: `("a")in`, args: []interface{}{1}},
					{clause: `("a") in thing`, args: []interface{}{1, 2, 3}},
					{clause: `("a")in?`, orSeparator: true, args: []interface{}{4, 5, 6}},
				},
			},
			expect: ` WHERE ("a")in AND ("a") IN thing OR ("a") IN ($1,$2,$3)`,
			args:   []interface{}{1, 1, 2, 3, 4, 5, 6},
		},
		{
			q: Query{
				in: []in{
					{clause: `("a")in?`, orSeparator: true, args: []interface{}{4, 5, 6}},
					{clause: `("a") in thing`, args: []interface{}{1, 2, 3}},
					{clause: `("a")in`, args: []interface{}{1}},
				},
			},
			expect: ` WHERE ("a") IN ($1,$2,$3) AND ("a") IN thing AND ("a")in`,
			args:   []interface{}{4, 5, 6, 1, 2, 3, 1},
		},
		{
			q: Query{
				in: []in{
					{clause: `("a")in?`, orSeparator: true, args: []interface{}{4, 5, 6}},
					{clause: `("a")in`, args: []interface{}{1}},
					{clause: `("a") in thing`, args: []interface{}{1, 2, 3}},
				},
			},
			expect: ` WHERE ("a") IN ($1,$2,$3) AND ("a")in AND ("a") IN thing`,
			args:   []interface{}{4, 5, 6, 1, 1, 2, 3},
		},
	}

	for i, test := range tests {
		test.q.dialect = &drivers.Dialect{LQ: '"', RQ: '"', UseIndexPlaceholders: true}
		result, args := inClause(&test.q, 1)
		if result != test.expect {
			t.Errorf("%d) Mismatch between expect and result:\n%s\n%s\n", i, test.expect, result)
		}
		if !reflect.DeepEqual(args, test.args) {
			t.Errorf("%d) Mismatch between expected args:\n%#v\n%#v\n", i, test.args, args)
		}
	}
}

func TestConvertQuestionMarks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		clause string
		start  int
		expect string
		count  int
	}{
		{clause: "hello friend", start: 1, expect: "hello friend", count: 0},
		{clause: "thing=?", start: 2, expect: "thing=$2", count: 1},
		{clause: "thing=? and stuff=? and happy=?", start: 2, expect: "thing=$2 and stuff=$3 and happy=$4", count: 3},
		{clause: `thing \? stuff`, start: 2, expect: `thing ? stuff`, count: 0},
		{clause: `thing \? stuff and happy \? fun`, start: 2, expect: `thing ? stuff and happy ? fun`, count: 0},
		{
			clause: `thing \? stuff ? happy \? and mad ? fun \? \? \?`,
			start:  2,
			expect: `thing ? stuff $2 happy ? and mad $3 fun ? ? ?`,
			count:  2,
		},
		{
			clause: `thing ? stuff ? happy \? fun \? ? ?`,
			start:  1,
			expect: `thing $1 stuff $2 happy ? fun ? $3 $4`,
			count:  4,
		},
		{clause: `?`, start: 1, expect: `$1`, count: 1},
		{clause: `???`, start: 1, expect: `$1$2$3`, count: 3},
		{clause: `\?`, start: 1, expect: `?`},
		{clause: `\?\?\?`, start: 1, expect: `???`},
		{clause: `\??\??\??`, start: 1, expect: `?$1?$2?$3`, count: 3},
		{clause: `?\??\??\?`, start: 1, expect: `$1?$2?$3?`, count: 3},
	}

	for i, test := range tests {
		res, count := convertQuestionMarks(test.clause, test.start)
		if res != test.expect {
			t.Errorf("%d) Mismatch between expect and result:\n%s\n%s\n", i, test.expect, res)
		}
		if count != test.count {
			t.Errorf("%d) Expected count %d, got %d", i, test.count, count)
		}
	}
}

func TestConvertInQuestionMarks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		clause string
		start  int
		group  int
		total  int
		expect string
	}{
		{clause: "?", expect: "(($1,$2,$3),($4,$5,$6),($7,$8,$9))", start: 1, total: 9, group: 3},
		{clause: "?", expect: "(($2,$3),($4))", start: 2, total: 3, group: 2},
		{clause: "hello friend", start: 1, expect: "hello friend", total: 0, group: 1},
		{clause: "thing ? thing", start: 2, expect: "thing ($2,$3) thing", total: 2, group: 1},
		{clause: "thing?thing", start: 2, expect: "thing($2)thing", total: 1, group: 1},
		{clause: `thing \? stuff`, start: 2, expect: `thing ? stuff`, total: 0, group: 1},
		{clause: `thing \? stuff and happy \? fun`, start: 2, expect: `thing ? stuff and happy ? fun`, total: 0, group: 1},
		{clause: "thing ? thing ? thing", start: 1, expect: "thing ($1,$2,$3) thing ? thing", total: 3, group: 1},
		{clause: `?`, start: 1, expect: `($1)`, total: 1, group: 1},
		{clause: `???`, start: 1, expect: `($1,$2,$3)??`, total: 3, group: 1},
		{clause: `\?`, start: 1, expect: `?`, total: 0, group: 1},
		{clause: `\?\?\?`, start: 1, expect: `???`, total: 0, group: 1},
		{clause: `\??\??\??`, start: 1, expect: `?($1,$2,$3)????`, total: 3, group: 1},
		{clause: `?\??\??\?`, start: 1, expect: `($1,$2,$3)?????`, total: 3, group: 1},
	}

	for i, test := range tests {
		res, count := convertInQuestionMarks(true, test.clause, test.start, test.group, test.total)
		if res != test.expect {
			t.Errorf("%d) Mismatch between expect and result:\n%s\n%s\n", i, test.expect, res)
		}
		if count != test.total {
			t.Errorf("%d) Expected %d, got %d", i, test.total, count)
		}
	}

	res, count := convertInQuestionMarks(false, "?", 1, 3, 9)
	if res != "((?,?,?),(?,?,?),(?,?,?))" {
		t.Errorf("Mismatch between expected and result: %s", res)
	}
	if count != 9 {
		t.Errorf("Expected 9 results, got %d", count)
	}
}

func TestWriteAsStatements(t *testing.T) {
	t.Parallel()

	query := Query{
		selectCols: []string{
			`a`,
			`a.fun`,
			`"b"."fun"`,
			`"b".fun`,
			`b."fun"`,
			`a.clown.run`,
			`COUNT(a)`,
		},
		dialect: &drivers.Dialect{LQ: '"', RQ: '"', UseIndexPlaceholders: true},
	}

	expect := []string{
		`"a"`,
		`"a"."fun" as "a.fun"`,
		`"b"."fun" as "b.fun"`,
		`"b"."fun" as "b.fun"`,
		`"b"."fun" as "b.fun"`,
		`"a"."clown"."run" as "a.clown.run"`,
		`COUNT(a)`,
	}

	gots := writeAsStatements(&query)

	for i, got := range gots {
		if expect[i] != got {
			t.Errorf(`%d) want: %s, got: %s`, i, expect[i], got)
		}
	}
}
