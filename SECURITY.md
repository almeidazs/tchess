# Segurança

## Versões com suporte

Só a versão mais recente recebe correções de segurança. Antes de relatar, confira se o problema acontece nela (`tchess --version`).

## Como relatar

Não abra uma issue pública. Use o relato privado do GitHub, em [Security → Report a vulnerability](https://github.com/almeidazs/tchess/security/advisories/new), e conte:

- a versão do tchess e o sistema operacional;
- os passos para reproduzir, com o arquivo PGN ou a config envolvida, se houver;
- o que um atacante consegue fazer com isso.

A primeira resposta sai em até 7 dias. Confirmado o problema, a correção vai numa nova versão e o relato é publicado junto, com crédito para quem encontrou, se a pessoa quiser.

## O que conta como vulnerabilidade

O tchess não usa a rede. O que ele faz fora do terminal é:

- ler arquivos PGN escolhidos por quem joga (pelo menu ou por `--pgn`);
- ler e gravar a config, as partidas e os sons em `~/.tchess`;
- chamar o player de áudio do sistema (`pw-play`, `paplay`, `aplay`, `afplay` ou PowerShell) passando o caminho de um arquivo WAV que ele mesmo gerou.

Interessam principalmente um PGN que faça o programa travar, consumir memória sem limite ou fechar; uma gravação fora de `~/.tchess`; e qualquer forma de executar comandos pelo caminho de um arquivo. Problemas de regras de xadrez ou da interface são bugs comuns e podem ir direto para as [issues](https://github.com/almeidazs/tchess/issues).
