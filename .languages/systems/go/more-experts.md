# Influential Go Experts and Their Contributions (with Links)

Below is a curated list of significant Go (Golang) experts, their main contributions, and representative resources. Each entry includes **valid links** to talks, design docs, books, or code.

---

## Core Language Creators & Stewards
- **Rob Pike** — *Co-creator; design & culture*
  - [Go at Google: Language Design in the Service of Software Engineering](https://go.dev/talks/2012/splash.article)  
  - [The First Go Program](https://go.dev/blog/first-go-program)

- **Robert Griesemer** — *Co-creator; generics contributor*
  - [Design: Additions to `go/types` to support type parameters](https://go.googlesource.com/proposal/+/master/design/47916-parameterized-go-types.md)

- **Ken Thompson** — *Co-creator; assembler & toolchain lineage*
  - [The Design of the Go Assembler (slides)](https://go.dev/talks/2016/asm.slide)  
  - [Why did we create Go? (Google I/O 2012 video)](https://www.youtube.com/watch?v=c-P5R0aMylM)

- **Russ Cox** — *Long-time technical lead; modules & language evolution*
  - [Secure Randomness in Go 1.22 (with Filippo Valsorda)](https://go.dev/blog/secure-randomness)  
  - [Versioning in Go ("vgo" essay series)](https://research.swtch.com/vgo)

- **Ian Lance Taylor** — *Generics design; gccgo compiler*
  - [Proposal: Adding Generics to Go](https://go.dev/blog/generics-proposal)

- **Bryan C. Mills (bcmills)** — *Modules; lazy loading; release engineering*
  - [Lazy Module Loading design doc](https://go.googlesource.com/proposal/+/master/design/36460-lazy-module-loading.md)  
  - [GitHub profile](https://github.com/bcmills)

- **Andrew Gerrand** — *Early Go advocate; Tour of Go co-author*
  - [A Tour of Go](https://go.dev/tour/)  
  - [GopherCon 2015 closing keynote](https://www.youtube.com/watch?v=0ht89TxZZnk)

---

## Compiler, Runtime & Performance
- **Austin Clements** — *Runtime & ABI modernization*
  - [Register-based Go calling convention](https://go.googlesource.com/proposal/+/master/design/40724-register-calling.md)  
  - [Go 1.17 Release Notes (register ABI intro)](https://go.dev/doc/go1.17)

- **Keith Randall, David Chase, Dan Scales, Cherry Zhang** — *Compiler back-end & performance (SSA, reg-ABI)*
  - [Go Issue #40724 – Switch to register calling convention](https://github.com/golang/go/issues/40724)

- **Michael Knyszek** — *Garbage Collector & runtime scalability*
  - [Proposals index (runtime metrics, page allocator, GC advances)](https://golang-id.org/proposal/)

- **Dmitry Vyukov** — *Race detector; scheduler design*
  - [Introducing the Go Race Detector](https://go.dev/blog/race-detector)  
  - [ThreadSanitizer at Google](https://testing.googleblog.com/2014/06/threadsanitizer-slaughtering-data-races.html)  
  - [Go’s work-stealing scheduler (explainer by Jaana Dogan)](https://rakyll.org/scheduler/)

---

## Security & Cryptography
- **Filippo Valsorda** — *x/crypto; modern crypto tooling*
  - [age (modern encryption tool)](https://filippo.io/age/)  
  - [Secure Randomness in Go 1.22 (with Russ Cox)](https://go.dev/blog/secure-randomness)

---

## Tooling & Analysis
- **Dominik Honnef** — *Staticcheck linter & tools*
  - [Staticcheck official site](https://staticcheck.dev/)  
  - [Staticcheck GitHub repo](https://github.com/dominikh/go-tools)

---

## Networking, Protocols & Servers
- **Lucas Clemente & Marten Seemann** — *HTTP/3 & QUIC in Go*
  - [quic-go GitHub repo](https://github.com/lucas-clemente/quic-go)

- **Matt Holt** — *Caddy web server (Automatic HTTPS)*
  - [Caddy GitHub repo](https://github.com/caddyserver/caddy)

- **Brad Fitzpatrick** — *Long-time stdlib/net contributor; http/2; infra*
  - [Wired profile on early Go at Google](https://www.wired.com/2013/07/gopher/)  
  - [x/net/http2 package docs](https://pkg.go.dev/golang.org/x/net/http2)

---

## Libraries, Frameworks & Platforms
- **Peter Bourgon** — *Go kit (microservices toolkit)*
  - [Go kit GitHub repo](https://github.com/go-kit/kit)  
  - [Go kit site](https://gokit.io/)

- **Steve Francia (spf13)** — *CLI, config, static sites*
  - [Cobra CLI framework](https://github.com/spf13/cobra)  
  - [Viper config library](https://github.com/spf13/viper)  
  - [Hugo static site generator](https://github.com/gohugoio/hugo)

- **Ben Johnson** — *BoltDB (embedded KV)*
  - [BoltDB repo](https://github.com/boltdb/bolt)  
  - [bbolt (maintained successor)](https://github.com/etcd-io/bbolt)

- **Julius Volz & Prometheus Team** — *Prometheus monitoring & TSDB*
  - [Prometheus GitHub]()
