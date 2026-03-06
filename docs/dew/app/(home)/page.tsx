import Link from 'next/link';
import Image from 'next/image';
import heroBg from '@/assets/hero_background.png';
import { DynamicCodeBlock } from 'fumadocs-ui/components/dynamic-codeblock';

const codeExample = `// Define once
var Users = dew.DefineSchema("users", dew.PostgreSQLDialect{}, func(t dew.Table[User]) struct {
    dew.Table[User]
    ID   dew.IntColumn
    Name dew.StringColumn
    Age  dew.IntColumn
} {
    return struct {
        dew.Table[User]
        ID   dew.IntColumn
        Name dew.StringColumn
        Age  dew.IntColumn
    }{
        Table: t,
        ID:    t.IntColumn("id"),
        Name:  t.StringColumn("name"),
        Age:   t.IntColumn("age"),
    }
})

// Query anywhere — type-safe, composable, no repo needed
adults, err := Users.From(db).
    Where(Users.Age.Gte(18)).
    OrderBy(dew.Desc(Users.Name)).
    Limit(10).
    All(ctx)`;

const beforeSQL = `rows, err := db.QueryContext(ctx,
    "SELECT * FROM users WHERE age >= $1 AND role = $2 ORDER BY name LIMIT $3",
    18, "admin", 10,
)
if err != nil {
    return nil, err
}
defer rows.Close()

var users []User
for rows.Next() {
    var u User
    err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.Role)
    if err != nil {
        return nil, err
    }
    users = append(users, u)
}`;

const afterDew = `users, err := Users.From(db).
    Where(Users.Age.Gte(18), Users.Role.Eq("admin")).
    OrderBy(dew.Asc(Users.Name)).
    Limit(10).
    All(ctx)`;

const comparisons = [
  {
    name: 'GORM',
    desc: 'Full ORM with auto-migrations, hooks, associations. Heavy abstraction — magic behind the scenes, hard to debug generated SQL, interface{} everywhere.',
    dew: 'Dew is just a query builder. No magic, no auto-migration, no hooks. You see exactly what SQL runs. Type-safe columns instead of string-based field names.',
  },
  {
    name: 'sqlx',
    desc: 'Thin wrapper over database/sql with struct scanning. Great, but queries are raw strings — no compile-time safety, no composability.',
    dew: 'Dew gives you the same directness as sqlx but with typed columns and composable builders. Refactor a column name and the compiler catches every usage.',
  },
  {
    name: 'ent',
    desc: 'Code-generated type-safe ORM. Powerful, but requires a build step, generates thousands of lines, and owns your schema.',
    dew: 'Dew achieves type safety through generics alone — zero code generation. Your schema is plain Go code. No build step, no generated files.',
  },
];

export default function HomePage() {
  return (
    <div className="flex flex-col">
      {/* Hero */}
      <section className="relative flex flex-col items-center justify-center text-center min-h-[80vh] overflow-hidden">
        <Image
          src={heroBg}
          alt=""
          fill
          className="object-cover -z-10 opacity-30"
          priority
        />
        <div className="relative z-10 px-4 py-20">
          <h1 className="text-5xl font-bold mb-4 tracking-tight">Dew</h1>
          <p className="text-xl text-muted-foreground mb-8 max-w-lg mx-auto">
            A lightweight, type-safe query builder for Go.<br />
            No ORM magic, no repo layers — just queries.
          </p>
          <div className="flex gap-4 justify-center">
            <Link href="/docs" className="inline-flex items-center justify-center px-6 py-2.5 text-sm font-medium rounded-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors">
              Get Started
            </Link>
            <Link href="https://github.com/dr3dnought/dew" className="inline-flex items-center justify-center px-6 py-2.5 text-sm font-medium rounded-full border bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors">
              GitHub
            </Link>
          </div>
          <div className="mt-6 max-w-sm mx-auto">
            <DynamicCodeBlock lang="bash" code="go get github.com/dr3dnought/dew" />
          </div>
        </div>
      </section>

      {/* Code Example */}
      <section className="px-4 py-20 max-w-4xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-3">Write queries, not abstractions</h2>
        <p className="text-center text-muted-foreground mb-10 max-w-2xl mx-auto">
          Define your schema once with typed columns. Query anywhere with full compile-time safety — no strings, no guessing.
        </p>
        <DynamicCodeBlock lang="go" code={codeExample} />
      </section>

      {/* Before/After */}
      <section className="px-4 py-20 max-w-3xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-3">Before and after</h2>
        <p className="text-center text-muted-foreground mb-10">
          Same query. Less code. Type-safe. No manual scanning.
        </p>
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Raw database/sql — 14 lines</p>
        <DynamicCodeBlock lang="go" code={beforeSQL} />
        <div className="flex items-center justify-center py-4 text-muted-foreground">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
        </div>
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">With dew — 5 lines</p>
        <DynamicCodeBlock lang="go" code={afterDew} />
        <p className="text-center text-sm text-muted-foreground mt-6">
          The compiler checks every column name.
        </p>
      </section>

      {/* Why not X? */}
      <section className="px-4 py-20 max-w-4xl mx-auto w-full">
        <h2 className="text-3xl font-bold text-center mb-3">Why dew?</h2>
        <p className="text-center text-muted-foreground mb-10">
          There are great tools out there. Here&#39;s where dew fits.
        </p>
        <div className="rounded-lg border bg-card overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-6 py-3 font-medium text-muted-foreground">Tool</th>
                <th className="text-left px-6 py-3 font-medium text-muted-foreground">Approach</th>
                <th className="text-left px-6 py-3 font-medium text-muted-foreground">Dew</th>
              </tr>
            </thead>
            <tbody>
              {comparisons.map((c) => (
                <tr key={c.name} className="border-b last:border-b-0">
                  <td className="px-6 py-4 font-semibold whitespace-nowrap align-top">{c.name}</td>
                  <td className="px-6 py-4 text-muted-foreground align-top">{c.desc}</td>
                  <td className="px-6 py-4 align-top">{c.dew}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* CTA */}
      <section className="px-4 py-16 text-center border-t">
        <h2 className="text-2xl font-bold mb-3">Ready to drop the boilerplate?</h2>
        <p className="text-muted-foreground mb-6">Get started in under 5 minutes.</p>
        <Link href="/docs" className="inline-flex items-center justify-center px-6 py-2.5 text-sm font-medium rounded-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors">
          Read the Docs
        </Link>
      </section>
    </div>
  );
}
