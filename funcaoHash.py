import pandas as pd
"""
Implementação Jenkins One-at-a-time Hash (https://en.wikipedia.org/wiki/Jenkins_hash_function)
Esta função é focada em bit mixing, com passos extras para gerar efeito avalanche, ou seja, qualquer diferença pequena se propagará mais e mais alterando o valor final.
Retornará um inteiro para cada palavra, o inteiro será de grande magnitude, possibilitado aplicar os módulos diversas vezes, optei por mesclar multiplicação por primo
e módulo a base de potência de dois.
"""


def jenkins_one_at_a_time(palavra: str) -> int:
    hash_value = 0
    for caractere in palavra:
        # ord() retorna o valor ASCII de cada caractere, ou a gente tem que fazer isso manualmente também?
        hash_value += ord(caractere)
        hash_value += (hash_value << 10)
        hash_value ^= (hash_value >> 6)

    hash_value += (hash_value << 3)
    hash_value ^= (hash_value >> 11)
    hash_value += (hash_value << 15)

    hash_value = (hash_value * 35969) % (2**32)

    # Essa função é orientada a 32 bits, isso vai limitar o valor do hash para 32 bits no Python, caso contrário, valores absurdamente grandes ocorrerão.
    return hash_value & 0xffffffff


# Caso queira testar, lembrar de colocar na mesma pasta do arquivo.
local_arq = "words.txt"

with open(local_arq, "r", encoding="utf-8") as arquivo:
    lista = pd.Series([linha.strip() for linha in arquivo])

# Tava dando erro na leitura, isso garante que qualquer valor nulo será ignorado e todos os valores de linha serão convertidos para string.
lista = lista.dropna().astype(str)
lista_hash = lista.apply(jenkins_one_at_a_time)

# Controle, apenas para verificar os valores obtidos, se obteve algum valor extremamente baixo ou alto.
m = lista_hash.nsmallest(100)
# m = lista_hash.nlargest(100)
for i in m:
    print(i)

# Verificando quantas possíveis colisões teríamos até aqui.
duplicatas = lista_hash.value_counts()
print(duplicatas[duplicatas > 1])
