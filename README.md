<div align="center">

# ♞ TChess

Xadrez para dois jogadores no terminal, escrito em Go.<br>
As peças são pixel art, dá para arrastá-las com o mouse, o tabuleiro gira para quem vai jogar<br>
e todas as regras da FIDE valem, de en passant à regra dos 75 lances.

[Instalação](#instalação) ·
[Como jogar](#como-jogar) ·
[Como funciona](#como-funciona) ·
[Correção e desempenho](#correção-e-desempenho) ·
[Desenvolvimento](#desenvolvimento)

<br>

<img src="assets/demo.gif" alt="Partida no terminal: peças arrastadas com o mouse e o tabuleiro girando para o próximo jogador" width="900">

</div>

## O que tem

- Regras completas: roque (que não pode passar por casa atacada), en passant (inclusive quando ele descobriria um xeque na horizontal), promoção com escolha da peça, cravadas, xeque duplo, mate e afogamento.
- Todos os jeitos de uma partida terminar. Material insuficiente, repetição quíntupla e a regra dos 75 lances encerram sozinhos. Repetição tripla e a regra dos 50 lances podem ser reivindicadas. Também há acordo, desistência e queda de seta.
- Peças desenhadas com meios-blocos `▀`, em três tamanhos que acompanham o tamanho do terminal.
- Mouse: arraste a peça até a casa, ou clique na origem e depois no destino. Os lances legais aparecem como pontos e as capturas como cantos marcados.
- Teclado: as setas param só onde existe lance possível, ou seja, nas suas peças que podem andar ou nos destinos da peça selecionada.
- Também dá para digitar o lance depois de `:`, em SAN ou em coordenadas (`Nf3`, `exd5`, `O-O`, `e2e4`). Enquanto você digita, as peças que o texto pode indicar acendem no tabuleiro.
- Depois de cada lance a peça desliza e o tabuleiro gira 180°. Qualquer tecla pula a animação, e o relógio do adversário só começa quando ela acaba. Quem prefere ver sempre do lado das brancas desliga o giro no menu ou com `f`.
- Relógio com incremento Fischer (bullet, blitz, rápida ou um tempo à sua escolha) e o tempo gasto em cada lance, com barras no histórico.
- Ao lado de cada jogador ficam as peças que ele já capturou e a vantagem de material.
- Dá para voltar um lance, se o adversário aceitar, e rever qualquer posição anterior com `[` e `]` ou clicando no histórico, sem parar o relógio.
- A partida em andamento fica salva sozinha. Se você fechar o terminal, o menu oferece continuar de onde parou, com os relógios como estavam.
- Toda partida terminada fica salva em PGN, com o tempo de cada lance nos comentários `[%clk]` e `[%emt]` que o Lichess lê. Dá para abrir qualquer uma de novo pelo menu, ou importar um PGN de outro programa.
- Seis temas de tabuleiro, escolhidos no menu com uma prévia do tabuleiro ou trocados com `s` durante a partida. Sons para lance, captura, xeque, roque, promoção e fim de partida, e uma ajuda com todos os atalhos no `?`. Tudo fica salvo para a próxima vez.
- Por baixo há um motor de bitboards que gera lances sem alocar memória e passa de 160 milhões de posições por segundo no perft.

<div align="center">
  <img src="assets/screenshot.png" alt="Partida em andamento com o cavalo selecionado, lances legais, peças capturadas e tempos por lance" width="900">
  <br>
  <sub>Ataque Møller depois de <code>12...Bxg5</code>, com o cavalo de f3 selecionado e <code>Nxg5</code> marcado como captura.</sub>
</div>

## Instalação

Com Go 1.27 ou mais recente:

```sh
go install github.com/almeidazs/tchess/cmd/tchess@latest
```

Com Docker, sem instalar nada (o volume guarda as preferências e as partidas):

```sh
docker run --rm -it -v tchess:/root/.tchess ghcr.io/almeidazs/tchess
```

A página de [releases](https://github.com/almeidazs/tchess/releases) tem binários para Linux, macOS e Windows (amd64 e arm64) e pacotes `.deb`, `.rpm`, `.apk` e `.pkg.tar.zst`:

```sh
sudo apt install ./tchess_*_amd64.deb       # Debian, Ubuntu
sudo dnf install ./tchess-*.x86_64.rpm      # Fedora
sudo apk add --allow-untrusted ./tchess_*.apk   # Alpine
sudo pacman -U ./tchess-*.pkg.tar.zst       # Arch
```

Para compilar a partir do código:

```sh
git clone https://github.com/almeidazs/tchess.git
cd tchess
make run
```

O jogo precisa de um terminal com truecolor e mouse, como Kitty, WezTerm, Ghostty, Alacritty, GNOME Terminal, iTerm2 ou Windows Terminal. Quanto maior a janela, maiores as peças.

## Como jogar

```sh
tchess                  # abre o menu
tchess --time 5+3       # começa direto: 5 minutos + 3 segundos por lance
tchess --time none      # sem relógio
tchess --pgn jogo.pgn   # abre uma partida de um arquivo PGN
tchess --theme Oceano   # outro tema só nesta execução
tchess --no-sound       # sem som só nesta execução
tchess perft 6          # autoteste do gerador de lances
tchess --version
```

| Ação | Mouse | Teclado |
| --- | --- | --- |
| Mover uma peça | arrastar até a casa, ou clicar na origem e no destino | `←↑↓→` ou `hjkl`, depois `enter` |
| Digitar o lance | | `:`, o lance e `enter` |
| Cancelar a seleção | clicar fora ou na mesma peça | `esc` |
| Escolher a promoção | clicar na peça do seletor | `↑↓` e `enter`, ou `q` `r` `b` `n` |
| Pedir para voltar o lance | | `u` (o adversário responde `y` ou `n`); na revisão, volta até a posição mostrada |
| Rever lances anteriores | clicar no lance do histórico | `[` e `]` (no fim da partida, também `←` e `→`), `home` e `end`, `esc` volta |
| Esconder ou mostrar o resultado | | `esc` no fim da partida |
| Exportar a partida em PGN | | `e` |
| Oferecer ou reivindicar empate | | `d` |
| Desistir | | `r` |
| Trocar o tema | | `s` |
| Ligar ou desligar o som | | `m` |
| Ligar ou desligar o giro do tabuleiro | | `f` |
| Ver os atalhos | | `?` |
| Sair | | `q` ou `ctrl+c` |

Quando a posição permite reivindicar empate por repetição tripla ou pela regra dos 50 lances, a barra de status avisa, e `d` faz a reivindicação. Fora desses casos, `d` oferece empate e o adversário responde com `y` ou `n`.

O menu tem os controles de tempo e, embaixo, as opções: o tema (abre a lista com uma prévia do tabuleiro em cada um), o giro do tabuleiro e "Abrir PGN". Quando há uma partida interrompida, o primeiro item é "Continuar".

Quem pede para voltar o lance é quem acabou de jogar, e o adversário decide. Para voltar mais, é só ir até a posição desejada na revisão e apertar `u`: o pedido é para voltar a partida até ali, feito por quem joga naquela posição. O relógio de quem pediu volta ao que marcava no momento do lance, sem o incremento que ganhou com ele, e o do adversário volta ao que marcava antes de começar a correr. Na revisão o tabuleiro mostra a posição escolhida, o relógio continua correndo e só dá para jogar de novo depois de voltar ao presente.

O campo do `:` aceita o lance escrito de vários jeitos: `Nf3`, `nf3`, `Ngf3`, `g1f3` ou `g1-f3`, roque como `O-O` ou `0-0`, captura com ou sem `x` e promoção como `e8=Q`, `e8Q` ou `e7e8q`. Sinais como `+`, `#`, `!` e `?` são ignorados. Se o texto serve para dois lances (`Nd2` com os dois cavalos podendo ir), o jogo pede para desfazer a ambiguidade. Escrever só `e8` abre o seletor de promoção. Em `bc3` com letra minúscula, que pode ser peão ou bispo, vale o peão, como na maioria dos aplicativos.

O tamanho das peças depende do terminal:

| Terminal | Peças |
| --- | --- |
| a partir de 122×46 | 11×10 pixels |
| a partir de 106×38 | 9×8 pixels |
| a partir de 90×30 | 7×6 pixels |
| a partir de 58×14 | um caractere Unicode por casa |

### Preferências

O tema, o som, o giro do tabuleiro e o último controle de tempo jogado ficam em `~/.tchess/config.toml`, que o jogo grava sozinho quando algo muda. O menu já abre no último ritmo. As flags `--theme` e `--no-sound` valem só para aquela execução e não mudam o arquivo.

```toml
theme = "Floresta"
sound = true
flip = true
time_control = "5+3"
```

Os temas são Madeira, Floresta, Oceano, Grafite, Coral e Mármore. Os sons são gerados pelo próprio programa e tocados pelo player do sistema: `pw-play`, `paplay` ou `aplay` no Linux, `afplay` no macOS e PowerShell no Windows. Os arquivos WAV ficam em `~/.tchess/sounds/`. Sem nenhum desses players, o jogo roda mudo.

### Partidas salvas e PGN

A partida em andamento é gravada em `~/.tchess/current.pgn` a cada lance, para poder ser retomada. Ao continuar, o relógio de quem está na vez volta a correr na hora. Começar outra partida descarta essa.

Toda partida que termina (mate, desistência, empate ou tempo) vai sozinha para `~/.tchess/games/`, num arquivo com a data e a hora, e sai do `current.pgn`. No meio do jogo, `e` salva uma cópia na mesma pasta. Cada lance leva o relógio depois dele (`[%clk 0:04:58.0]`) e o tempo gasto (`[%emt 0:00:04.2]`), então o arquivo pode ser importado direto no Lichess. "Abrir PGN", no menu, lista essa pasta e aceita qualquer outro caminho; `--pgn` faz o mesmo pela linha de comando. Um arquivo com várias partidas, como uma exportação do Lichess ou um torneio, mostra a lista delas com os jogadores, o resultado e o número de lances, e as que não puderem ser lidas são puladas com um aviso. A leitura aceita o que outros programas escrevem: comentários, variantes (que são ignoradas), NAGs, posição inicial por FEN e lances em SAN ou em coordenadas.

Uma partida terminada abre na posição final, com o resultado e o motivo na barra de status, e as setas passam pelos lances. Uma inacabada pergunta antes como continuar: com os relógios do próprio PGN, quando ele traz `[%clk]`, com um dos ritmos do menu (os dois relógios cheios) ou sem relógio.

## Como funciona

```mermaid
flowchart LR
    main["cmd/tchess<br/>flags · perft"] --> model

    subgraph tui["tui/ · Bubble Tea"]
        model["Model<br/>Update · View"] --> board["board.go<br/>framebuffer ▀ + sprites"]
        model --> anim["anim.go<br/>deslize + giro 180°"]
        model --> mouse["mouse.go<br/>arrastar e soltar"]
        model --> cursor["cursor.go<br/>navegação por setas"]
        model --> input["moveinput.go<br/>lance digitado"]
        model --> record["record.go · files.go<br/>PGN, retomar, exportar"]
    end

    subgraph internal["internal/"]
        config["config/<br/>~/.tchess/config.toml"]
        atomic["atomicfile/<br/>gravação atômica"]
        sound["sound/<br/>síntese + player do sistema"]
    end

    subgraph core["núcleo · sem dependência de UI"]
        chess["chess/<br/>regras · movegen · FEN · SAN · PGN · perft"]
        clock["clock/<br/>incremento Fischer"]
    end

    model --> chess
    model --> clock
    main --> config
    model --> sound
    record --> atomic
    config --> atomic
```

O pacote `chess` é o motor de regras, o `clock` é o relógio e o `tui` é a interface. A interface usa o motor por uma API pequena (`Game.Play`, `Game.Undo`, `Game.PositionAt`, `Game.PGN`, `ParsePGN` e poucas outras). O motor devolve só enums, e todo texto que aparece na tela mora na interface.

<details>
<summary><b>O motor: bitboards e lances legais</b></summary>

<br>

Cada tipo de peça de cada cor é guardado num `uint64`, com um bit por casa (`A1` é o bit 0). Ao lado fica um *mailbox* de 64 casas, sempre sincronizado, para responder em O(1) qual peça está em cada casa. Um lance cabe em 16 bits: 6 para a origem, 6 para o destino e 4 de flags (captura, promoção, roque, en passant e avanço duplo).

Os ataques de bispos e torres usam magic bitboards. As casas relevantes (`mask`) são multiplicadas por um número escolhido e deslocadas, e o resultado vira o índice de uma tabela de ataques pré-calculada:

```go
func (m *magic) attacks(occ Bitboard) Bitboard {
	return m.table[(uint64(occ&m.mask)*m.mul)>>m.shift]
}
```

Esses números são procurados na inicialização, com um gerador determinístico, e a busca leva poucos milissegundos.

O gerador produz direto só os lances legais, sem jogar cada lance para testar se o rei ficou em xeque:

- O rei só vai para casas que o adversário não ataca. Esses ataques são calculados com o rei fora do tabuleiro, para ele não recuar ao longo da linha do próprio xeque.
- Em xeque duplo, só o rei se move.
- Em xeque simples, as outras peças precisam capturar quem dá o xeque ou bloquear a linha. Isso é feito com uma máscara de xeque.
- Peças cravadas só andam sobre a linha da cravada.
- O en passant tira duas peças da mesma fileira de uma vez, então passa por um teste exato de xeque descoberto.

Jogar e desfazer um lance atualiza bitboards, mailbox e hash Zobrist de forma incremental, em cerca de 19 ns pelo par. O hash serve para reconhecer posições repetidas. Como manda a FIDE, a casa de en passant só entra no hash quando a captura é legal de verdade.

</details>

<details>
<summary><b>A interface: pixels num terminal</b></summary>

<br>

O caractere `▀` pinta a metade de cima da célula com a cor da frente e a metade de baixo com a cor de fundo. Com isso, cada célula do terminal mostra dois pixels independentes e quase quadrados:

```
uma célula            dois pixels
┌─────────┐           ┌─────────┐
│    ▀    │  fg  →    │█████████│  pixel de cima
│         │  bg  →    │▒▒▒▒▒▒▒▒▒│  pixel de baixo
└─────────┘           └─────────┘
```

O tabuleiro é desenhado primeiro num framebuffer de pixels (casas, destaques, pontos de lance, brilho do xeque e sprites) e depois convertido em texto. Os códigos de cor só são emitidos quando a cor muda, e um quadro inteiro sai em cerca de 0,4 ms. Esse framebuffer é genérico e fica no pacote `internal/pixel`.

As peças são pixel art escrita no próprio código. A largura ímpar dá a cada peça uma coluna central e um pixel de margem de cada lado:

```text
...........      .  transparente
....###....      #  contorno
...#ooo#...      o  corpo
...#ooo#...      x  detalhe
....###....
...#ooo#...
....#o#....
...#ooo#...
..#ooooo#..
.#########.
```

Como o tabuleiro fica igual depois de uma rotação de 180°, as casas não precisam se mexer na animação: só as peças giram em volta do centro. No meio do giro as posições encolhem um pouco (`1 − 0,3·sen θ`) para os cantos não saírem do tabuleiro, e as peças mais baixas são desenhadas por cima das outras, o que dá uma ideia de profundidade.

A posição de cada elemento na tela é uma função pura do tamanho do terminal. Por isso o clique do mouse usa a mesma geometria do desenho, inclusive com o tabuleiro virado.

</details>

<details>
<summary><b>Relógio e tempo por lance</b></summary>

<br>

O relógio (`clock/`) implementa incremento Fischer e recebe a fonte de tempo de fora, o que deixa os testes determinísticos. O tempo do adversário só começa a correr quando a animação de giro termina, para ninguém perder segundos vendo o tabuleiro virar. A mesma regra vale para o tempo de cada lance que aparece no histórico.

</details>

<details>
<summary><b>Som sem cgo</b></summary>

<br>

Tocar áudio direto do processo em Go exige cgo e ALSA no Linux, o que impediria os binários estáticos para todas as plataformas. Por isso o pacote `internal/sound` sintetiza cada efeito em código, com senoides com envelope e ruído filtrado: o "toc" de madeira de um lance, um impacto mais grave na captura, dois toques no roque, um sino no xeque, um arpejo na promoção e um acorde no fim da partida. O resultado vira WAV de 16 bits, é gravado uma vez em `~/.tchess/sounds/` e tocado por um processo do player do sistema, sem bloquear a interface.

</details>

## Correção e desempenho

O gerador de lances é conferido com perft, que conta todas as posições alcançáveis até uma profundidade e compara com os valores publicados no [Chessprogramming Wiki](https://www.chessprogramming.org/Perft_Results). Um erro em qualquer regra muda a contagem.

| Posição | Profundidade | Nós | Tempo¹ |
| --- | :---: | ---: | ---: |
| Inicial | 6 | 119.060.324 | 0,73 s |
| Kiwipete | 5 | 193.690.690 | 0,80 s |
| Posições 3 a 6 do wiki | 4 e 5 | conferidas no `go test` | |

| Benchmark¹ | Resultado |
| --- | ---: |
| Perft da posição inicial | cerca de 163 milhões de nós/s |
| Gerar os lances legais do Kiwipete (48 lances) | 126 ns, sem alocação |
| Jogar e desfazer um lance | 19 ns, sem alocação |
| Desenhar um quadro da interface | cerca de 0,4 ms |

<sub>¹ Intel Core i7-13620H, Go 1.27, uma thread. Dá para repetir com <code>make bench</code> e <code>make perft</code>.</sub>

Os testes também cobrem cada regra separadamente (en passant, roque, cravadas, promoção, SAN, PGN, desfazer e empates), o relógio e a interface, com cliques e arrastos no tabuleiro virado, navegação pelo teclado, promoção, animações e o desenho em todos os tamanhos de tela. O lance digitado, o voltar lance com os relógios, a revisão, exportar, importar e retomar uma partida, os seletores de tema e de PGN, a config e a escolha de cada som também têm testes, que gravam numa pasta temporária e não tocam áudio de verdade.

## Desenvolvimento

`make help` lista os comandos:

| Comando | O que faz |
| --- | --- |
| `make build` | compila em `bin/tchess` |
| `make run` | compila e abre o jogo |
| `make test` | roda os testes com o detector de race |
| `make test-short` | roda os testes com perft mais raso |
| `make bench` | benchmarks do motor e da interface |
| `make perft` | perft(6) da posição inicial |
| `make cover` | resumo de cobertura |
| `make lint`, `make fmt` | [golangci-lint](https://golangci-lint.run) com a configuração do repositório |
| `make assets` | gera de novo o print e o GIF deste README |
| `make snapshot` | gera os pacotes de release localmente com [GoReleaser](https://goreleaser.com) |

As versões saem com `git tag vX.Y.Z && git push --tags`: o GoReleaser gera os binários, os pacotes e a imagem Docker. O que muda em cada uma fica no [CHANGELOG](CHANGELOG.md). Falhas de segurança devem ser relatadas como explica o [SECURITY](SECURITY.md), e a convivência no projeto segue o [código de conduta](CODE_OF_CONDUCT.md).

As imagens deste README saem de [`tools/shot`](tools/shot). Ele controla a interface de verdade (teclas, cliques e arrastos) com um relógio falso e transforma a saída ANSI de cada quadro em imagem, então print e GIF sempre batem com o código.

```
.
├── chess/                 motor de regras (sem UI)
│   ├── types.go           Color, PieceType, Piece
│   ├── square.go          Square e coordenadas
│   ├── bitboard.go        Bitboard e máscaras
│   ├── move.go            Move (16 bits) e flags · movelist.go: buffer sem alocação
│   ├── castling.go        direitos de roque
│   ├── position.go        Position: estado e consultas de ataque
│   ├── makemove.go        Make / Unmake incrementais
│   ├── attacks.go         tabelas de ataque · magic.go: magic bitboards
│   ├── movegen.go         gerador de lances legais
│   ├── material.go        valores, material insuficiente
│   ├── zobrist.go         hash de posição
│   ├── fen.go · san.go    notações · pgn.go: escrita e leitura de PGN
│   ├── game.go            Game: histórico, lances, Undo e PositionAt
│   ├── gameover.go        mate, empates, repetição, desistência, queda de seta
│   ├── outcome.go         Outcome e Termination · errors.go
│   └── perft.go           contagem de nós
├── clock/                 TimeControl (timecontrol.go) e Clock (clock.go)
├── internal/
│   ├── pixel/             framebuffer de meios-blocos ▀, genérico
│   ├── atomicfile/        gravação atômica (arquivo temporário + rename)
│   ├── config/            leitura e gravação de ~/.tchess/config.toml
│   └── sound/             síntese dos efeitos (synth.go) e player do sistema (player.go)
├── tui/                   interface Bubble Tea
│   ├── types.go           tipos de estado · messages.go: mensagens e ticks
│   ├── model.go           Model, Init, Update
│   ├── game.go            fluxo da partida · keyboard.go · mouse.go · cursor.go · selection.go
│   ├── moveinput.go       lance digitado: leitura tolerante e destaque ao vivo
│   ├── undo.go            pedir e aceitar voltar o lance · review.go: rever posições anteriores
│   ├── record.go          partida ↔ PGN com relógios · files.go: retomar, exportar, listar · export.go
│   ├── prefs.go           tema, som, giro e gravação das preferências · help.go: a ajuda do ?
│   ├── layout.go          geometria adaptativa · coords.go: casa ↔ tela
│   ├── view.go            composição da tela · status.go · panel.go · format.go
│   ├── board.go           aparência das casas · board_pixel.go · board_glyph.go
│   ├── anim.go            deslize e giro do tabuleiro
│   ├── sprites.go         desenho dos sprites · sprite_data.go: a pixel art
│   ├── theme.go           os seis temas e os estilos · glyphs.go · text.go (textos em português)
│   ├── menu.go            menu: continuar, controles de tempo e opções
│   ├── themes.go          seletor de tema com prévia do tabuleiro
│   ├── openpgn.go         seletor de arquivos PGN e lista das partidas de um arquivo
│   └── clocks.go          escolha do ritmo para continuar uma partida importada
├── cmd/tchess/            binário (main.go, perft.go)
├── tools/shot/            gerador das imagens do README (módulo separado)
└── assets/                print e GIF gerados
```

Cada pacote tem um `doc.go` com a sua documentação. Tipos do domínio ficam em arquivos com o nome do tipo (`square.go`, `move.go`), e os tipos de estado da interface ficam juntos em `tui/types.go`. O comportamento de cada recurso fica no arquivo dele, e os testes seguem a mesma divisão (`movegen_test.go`, `drag_test.go`, `panel_test.go`).

## Licença

Distribuído sob a [Licença Apache 2.0](LICENSE). Copyright 2026 Almeida.
