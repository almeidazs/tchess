# Changelog

As mudanças de cada versão ficam aqui. O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e as versões seguem o [Versionamento Semântico](https://semver.org/lang/pt-BR/). As notas de cada release no GitHub são geradas a partir dos commits.

## [Não lançado]

Primeira versão.

### Adicionado

- Motor de regras com bitboards: lances legais, roque, en passant, promoção, xeque, mate e afogamento, conferido por perft.
- Todos os finais de partida da FIDE: material insuficiente, repetição quíntupla e regra dos 75 lances automáticas; repetição tripla e regra dos 50 lances reivindicáveis; acordo, desistência e queda de seta.
- Interface no terminal com peças em pixel art em três tamanhos, tabuleiro que gira para quem vai jogar (opcional) e animação de cada lance.
- Mouse (arrastar ou clicar) e teclado (setas que param só em lances possíveis), além de lance digitado depois de `:` em SAN ou coordenadas.
- Relógio com incremento Fischer, tempo gasto em cada lance e peças capturadas.
- Seis temas de tabuleiro com prévia no menu, sons gerados pelo próprio programa e ajuda com todos os atalhos no `?`.
- Voltar lances com o aceite do adversário, um de cada vez ou até uma posição escolhida na revisão.
- Revisão de qualquer posição anterior, pelo histórico ou com `[` e `]`.
- PGN: partidas terminadas salvas sozinhas, exportação com `[%clk]` e `[%emt]`, importação de arquivos com uma ou várias partidas e escolha do ritmo para continuar uma partida inacabada.
- Retomar a partida interrompida, com os relógios como estavam.
- Preferências em `~/.tchess/config.toml`.
- Pacotes deb, rpm, apk e Arch e imagem Docker em `ghcr.io/almeidazs/tchess`.

[Não lançado]: https://github.com/almeidazs/tchess/commits/main
