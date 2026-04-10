// SummaryTab — target intelligence overview (fake data for now)

const FAKE = {
  totalHosts: 247,
  uniqueIPs: 89,
  statusGroups: { '2xx': 156, '3xx': 43, '4xx': 38, '5xx': 10 },
  statusCodes: [
    { code: '200', count: 134 },
    { code: '201', count: 22 },
    { code: '301', count: 31 },
    { code: '302', count: 12 },
    { code: '401', count: 18 },
    { code: '403', count: 20 },
    { code: '500', count: 7 },
    { code: '503', count: 3 },
  ],
  techStack: [
    { name: 'Nginx',       count: 89 },
    { name: 'Cloudflare',  count: 67 },
    { name: 'React',       count: 45 },
    { name: 'jQuery',      count: 34 },
    { name: 'PHP',         count: 23 },
    { name: 'WordPress',   count: 18 },
    { name: 'Node.js',     count: 15 },
    { name: 'Apache',      count: 12 },
    { name: 'Bootstrap',   count: 9  },
    { name: 'Laravel',     count: 6  },
  ],
  cnames: [
    { cname: 'cloudfront.net',    count: 34 },
    { cname: 'amazonaws.com',     count: 23 },
    { cname: 'fastly.net',        count: 12 },
    { cname: 'cloudflare.com',    count: 8  },
    { cname: 'azurefd.net',       count: 4  },
    { cname: 'akamaiedge.net',    count: 3  },
  ],
  topIPs: [
    { ip: '104.21.14.x',   count: 12 },
    { ip: '172.67.183.x',  count: 8  },
    { ip: '13.227.85.x',   count: 7  },
    { ip: '151.101.x.x',   count: 6  },
    { ip: '52.84.x.x',     count: 5  },
  ],
  badges: [
    { badge: 'api',        count: 23, color: 'var(--accent)'  },
    { badge: 'auth',       count: 18, color: 'var(--orange)'  },
    { badge: 'cms',        count: 12, color: 'var(--yellow)'  },
    { badge: 'admin',      count: 7,  color: 'var(--red)'     },
    { badge: 'monitoring', count: 4,  color: 'var(--green)'   },
    { badge: 'dev',        count: 5,  color: 'var(--orange)'  },
    { badge: 'cicd',       count: 3,  color: 'var(--yellow)'  },
    { badge: 'storage',    count: 6,  color: 'var(--accent)'  },
  ],
  triageReviewed: 45,
  juicyHits: 31,
}

const scGroupColor: Record<string, string> = {
  '2xx': 'var(--green)',
  '3xx': 'var(--orange)',
  '4xx': 'var(--red)',
  '5xx': 'var(--yellow)',
}

function statusCodeColor(code: string): string {
  if (code.startsWith('2')) return 'var(--green)'
  if (code.startsWith('3')) return 'var(--orange)'
  if (code.startsWith('4')) return 'var(--red)'
  if (code.startsWith('5')) return 'var(--yellow)'
  return 'var(--text-dim)'
}

function Bar({ value, max, color }: { value: number; max: number; color: string }) {
  const pct = Math.round((value / max) * 100)
  return (
    <div style={{ flex: 1, height: 4, background: 'var(--border2)', borderRadius: 2, overflow: 'hidden' }}>
      <div style={{ width: `${pct}%`, height: '100%', background: color, borderRadius: 2, transition: 'width 0.3s' }} />
    </div>
  )
}

function Card({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      padding: '18px 20px',
      display: 'flex',
      flexDirection: 'column',
      gap: 14,
      ...style,
    }}>
      {children}
    </div>
  )
}

function CardTitle({ children }: { children: React.ReactNode }) {
  return (
    <div style={{
      fontSize: 11,
      fontWeight: 700,
      color: 'var(--text-muted)',
      textTransform: 'uppercase',
      letterSpacing: '0.08em',
      fontFamily: "'Fira Code', monospace",
      paddingBottom: 10,
      borderBottom: '1px solid var(--border)',
    }}>
      {children}
    </div>
  )
}

function StatCard({ label, value, color }: { label: string; value: number | string; color?: string }) {
  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      padding: '16px 20px',
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      gap: 6,
    }}>
      <div style={{ fontSize: 11, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.07em', fontFamily: "'Fira Code', monospace" }}>
        {label}
      </div>
      <div style={{ fontSize: 28, fontWeight: 700, color: color ?? 'var(--text)', lineHeight: 1 }}>
        {value}
      </div>
    </div>
  )
}

export default function SummaryTab({ domain }: { domain: string }) {
  const d = FAKE
  const maxTech = d.techStack[0]?.count ?? 1
  const maxCode = Math.max(...d.statusCodes.map(s => s.count))
  const maxCname = d.cnames[0]?.count ?? 1
  const maxIP = d.topIPs[0]?.count ?? 1
  const triagePct = Math.round((d.triageReviewed / d.totalHosts) * 100)

  return (
    <div style={{ flex: 1, overflow: 'auto', padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 18 }}>

      {/* Title */}
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 12 }}>
        <div style={{ fontSize: 15, fontWeight: 600 }}>Target Summary</div>
        <div style={{ fontSize: 12, color: 'var(--text-muted)', fontFamily: "'Fira Code', monospace" }}>{domain}</div>
        <div style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--text-muted)', fontStyle: 'italic' }}>
          ⚠ fake data — wire up backend to populate
        </div>
      </div>

      {/* Stat cards row */}
      <div style={{ display: 'flex', gap: 12 }}>
        <StatCard label="Total Hosts"  value={d.totalHosts} />
        <StatCard label="Unique IPs"   value={d.uniqueIPs} />
        <StatCard label="Juicy Hits"   value={d.juicyHits} color="var(--orange)" />
        {Object.entries(d.statusGroups).map(([grp, count]) => (
          <StatCard key={grp} label={grp} value={count} color={scGroupColor[grp]} />
        ))}
      </div>

      {/* Row 2: tech stack + status codes + badges */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>

        {/* Tech stack */}
        <Card>
          <CardTitle>Tech Stack</CardTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
            {d.techStack.map(({ name, count }) => (
              <div key={name} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 90, fontSize: 12, color: 'var(--text)', flexShrink: 0, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {name}
                </div>
                <Bar value={count} max={maxTech} color="var(--accent)" />
                <div style={{ width: 28, fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', fontFamily: "'Fira Code', monospace", flexShrink: 0 }}>
                  {count}
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Status code breakdown */}
        <Card>
          <CardTitle>Status Codes</CardTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
            {d.statusCodes.map(({ code, count }) => (
              <div key={code} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 36, fontSize: 12, fontFamily: "'Fira Code', monospace", color: statusCodeColor(code), flexShrink: 0 }}>
                  {code}
                </div>
                <Bar value={count} max={maxCode} color={statusCodeColor(code)} />
                <div style={{ width: 28, fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', fontFamily: "'Fira Code', monospace", flexShrink: 0 }}>
                  {count}
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Badges + triage */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Card style={{ flex: 1 }}>
            <CardTitle>Badges</CardTitle>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {d.badges.map(({ badge, count, color }) => (
                <div key={badge} style={{
                  display: 'flex', alignItems: 'center', gap: 6,
                  background: 'var(--surface2)', border: '1px solid var(--border2)',
                  borderRadius: 6, padding: '5px 10px',
                }}>
                  <span style={{ fontSize: 11, color, fontWeight: 600 }}>{badge}</span>
                  <span style={{ fontSize: 11, color: 'var(--text-muted)', fontFamily: "'Fira Code', monospace" }}>{count}</span>
                </div>
              ))}
            </div>
          </Card>

          <Card>
            <CardTitle>Triage Progress</CardTitle>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                  <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text)', fontFamily: "'Fira Code', monospace" }}>{d.triageReviewed}</span>
                  {' / '}{d.totalHosts} reviewed
                </span>
                <span style={{ fontSize: 12, color: 'var(--accent)', fontFamily: "'Fira Code', monospace", fontWeight: 700 }}>{triagePct}%</span>
              </div>
              <div style={{ height: 6, background: 'var(--border2)', borderRadius: 3, overflow: 'hidden' }}>
                <div style={{ width: `${triagePct}%`, height: '100%', background: 'var(--accent)', borderRadius: 3, transition: 'width 0.3s' }} />
              </div>
            </div>
          </Card>
        </div>
      </div>

      {/* Row 3: CNAMEs + IPs + ASN */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>

        {/* Top CNAMEs */}
        <Card>
          <CardTitle>Top CNAMEs</CardTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
            {d.cnames.map(({ cname, count }) => (
              <div key={cname} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ flex: 1, fontSize: 12, fontFamily: "'Fira Code', monospace", color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {cname}
                </div>
                <Bar value={count} max={maxCname} color="var(--orange)" />
                <div style={{ width: 28, fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', fontFamily: "'Fira Code', monospace", flexShrink: 0 }}>
                  {count}
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Top IPs */}
        <Card>
          <CardTitle>Top IPs</CardTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
            {d.topIPs.map(({ ip, count }) => (
              <div key={ip} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 110, fontSize: 12, fontFamily: "'Fira Code', monospace", color: 'var(--text)', flexShrink: 0 }}>
                  {ip}
                </div>
                <Bar value={count} max={maxIP} color="var(--yellow)" />
                <div style={{ width: 28, fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', fontFamily: "'Fira Code', monospace", flexShrink: 0 }}>
                  {count}
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* ASN — placeholder */}
        <Card>
          <CardTitle>ASN Intelligence</CardTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                placeholder="e.g. 13335"
                disabled
                style={{
                  flex: 1,
                  background: 'var(--surface2)',
                  border: '1px solid var(--border2)',
                  color: 'var(--text-muted)',
                  padding: '6px 10px',
                  borderRadius: 6,
                  fontSize: 12,
                  fontFamily: "'Fira Code', monospace",
                  cursor: 'not-allowed',
                }}
              />
              <button disabled style={{
                background: 'var(--surface2)',
                border: '1px solid var(--border2)',
                color: 'var(--text-muted)',
                padding: '6px 14px',
                borderRadius: 6,
                fontSize: 12,
                cursor: 'not-allowed',
              }}>
                Lookup
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[
                { label: 'AS13335',    value: 'Cloudflare, Inc.' },
                { label: 'Prefixes',   value: '12 IPv4 ranges' },
                { label: 'Total IPs',  value: '196,608' },
                { label: 'Peers',      value: '8 upstream ASNs' },
              ].map(({ label, value }) => (
                <div key={label} style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12 }}>
                  <span style={{ color: 'var(--text-muted)', fontFamily: "'Fira Code', monospace" }}>{label}</span>
                  <span style={{ color: 'var(--text)' }}>{value}</span>
                </div>
              ))}
            </div>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', fontStyle: 'italic', marginTop: 4 }}>
              ASN lookup not yet wired up
            </div>
          </div>
        </Card>

      </div>
    </div>
  )
}
