Självklart! Här är en färdig, ren och komplett **README.md** som du kan klistra in direkt i ditt repo.
Den beskriver exakt hur man kör ditt MapReduce-system med koordinator + workers + tester.

---

# MapReduce Lab 2 — How to Run Coordinator & Workers

Detta dokument beskriver hur du kör hela MapReduce-systemet (koordinator + workers), hur du bygger plugin-filer samt hur du kör testscriptet **test-mr.sh**.

Alla kommandon nedan körs från:

```
~/Desktop/Distributed_systems/LAB2/6.5840/src/main
```

---

## ⭐ 1. Bygg plugin (wc.so)

Varje gång du ändrar något i `mr/`-katalogen måste du bygga om plugin-filen.

```bash
rm -f wc.so
go build -race -buildmode=plugin ../mrapps/wc.go
```

> Flaggan `-race` måste matcha när du kör koordinators/worker med `-race`.

---

## ⭐ 2. Starta koordinatorn

Öppna en **ny terminal** och kör:

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
rm -f mr-out*
go run -race mrcoordinator.go pg-*.txt
```

Koordinatorn:

* startar RPC-servern
* delar ut map- och reduce-tasks
* avslutar av sig själv när allt arbete är klart

Låt denna terminal vara öppen.

---

## ⭐ 3. Starta en (eller flera) workers

Öppna **en annan terminal** och kör:

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
go run -race mrworker.go wc.so
```

Om allting fungerar korrekt:

* du **kommer inte** se felet
  `rpc: can't find method Coordinator.AssignTask`
* workern börjar direkt hämta tasks, köra map/reduce och rapportera tillbaka

Du kan starta flera workers i olika terminaler om du vill testa parallellism.

---

## ⭐ 4. Kontrollera resultatet

När koordinatorn avslutat (dvs den stänger av sig själv), kontrollera utskriftsfilerna:

```bash
ls mr-out-*
cat mr-out-* | sort | head
```

Du ska se ungefär:

```
A 509
ABOUT 2
ACT 8
...
```

Det betyder att MapReduce-jobbet lyckades.

---

## ⭐ 5. Kör hela testsviten

När du vet att manuell körning fungerar kör du:

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
bash test-mr.sh
```

Om allt är korrekt implementerat ska du se:

```
*** PASSED ALL TESTS
```

Testerna inkluderar:

* word count correctness
* indexer correctness
* map parallelism
* reduce parallelism
* worker crash recovery
* job counting
* early exit handling

---

## ⭐ 6. Tips

* Kom ihåg att **plugin-filer måste byggas om** varje gång du ändrar något i `mr/`.
* Om workern dör direkt → kontrol lera att wc.so är byggd med samma `-race` flagga.
* Om du får RPC-fel → koordinatorn kör troligen en gammal version. Starta om allt.

---

Vill du ha en ännu mer komplett README (med installation, strukturöversikt, förklaringar av RPC-API, exempelbilder), så fixar jag det också!


## MANUAL tests
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main

rm -f wc.so mr-out-* seq-out dist-out
go build -buildmode=plugin ../mrapps/wc.go

# sekventiellt
go run mrsequential.go wc.so pg-*.txt

# hela sorterade outputen (ingen head, ingen more)
cat mr-out-0 | sort > seq-out
2️⃣ Distribuerad output (dist-out)
bash
Kopiera kod
rm -f mr-out-*

# kör koordinator (i samma katalog)
go run mrcoordinator.go pg-*.txt
# (i en annan terminal:)
go run mrworker.go wc.so
När koordinatorn är klar:

bash
Kopiera kod
cat mr-out-* | sort > dist-out
3️⃣ Jämför – och få tystnad 😄
bash
Kopiera kod
diff seq-out dist-out