import numpy as np


class IndiceHash:
    """ Representa a estrutura de um índice hash usando buckets dinâmicos. """

    def __init__(self, num_buckets, tamanho_bucket):
        self.num_buckets = num_buckets
        self.tamanho_bucket = tamanho_bucket
        # Lista estática com listas dinâmicas de buckets
        self.tabela = [[Bucket(tamanho_bucket)] for _ in range(num_buckets)]
        self.mapeamento_chaves = {}  # Dicionário que mapeia chaves para números de página

    def funcao_hash(self, chave):
        hash_value = 0
        for caractere in chave:
            # ord() retorna o valor ASCII de cada caractere, ou a gente tem que fazer isso manualmente também?
            hash_value += ord(caractere)
            hash_value += (hash_value << 10)
            hash_value ^= (hash_value >> 6)

        hash_value += (hash_value << 3)
        hash_value ^= (hash_value >> 11)
        hash_value += (hash_value << 15)

        hash_value = (hash_value * 35969) % (2**32)

        # Essa função é orientada a 32 bits, isso vai limitar o valor do hash para 32 bits no Python, caso contrário, valores absurdamente grandes ocorrerão.
        return (hash_value & 0xffffffff) % 11

    def inserir(self, chave, num_pagina):
        """ Insere a chave no índice hash, associando-a a um número de página. """
        indice = self.funcao_hash(chave)
        for bucket in self.tabela[indice]:
            if bucket.adicionar_endereco(num_pagina):
                self.mapeamento_chaves[chave] = num_pagina
                return

        novo_bucket = Bucket(self.tamanho_bucket)
        novo_bucket.adicionar_endereco(num_pagina)
        self.tabela[indice].append(novo_bucket)
        self.mapeamento_chaves[chave] = num_pagina
        # print(f"Overflow! Novo bucket adicionado ao índice {indice}.")

    def buscar(self, chave):
        """ Busca uma chave no índice hash e retorna o número da página correspondente. """
        return self.mapeamento_chaves.get(chave, None)

    def exibir_tabela(self):
        """ Exibe o conteúdo do índice hash mostrando os buckets dinâmicos. """
        for i, lista_buckets in enumerate(self.tabela):
            print(f"Bucket {i}: ")
            for j, bucket in enumerate(lista_buckets):
                print(f"  Sub-bucket {j}: ", end="")
                bucket.exibir_conteudo()

        print("\n Mapeamento de Chaves -> Páginas")
        for chave, pagina in self.mapeamento_chaves.items():
            print(f"Chave '{chave}' -> Página {pagina}")


class Bucket:
    """ Representa um bucket de um índice hash, armazenando apenas endereços de páginas. """

    def __init__(self, tamanho_max):
        self.tamanho_max = tamanho_max
        # Inicializa com -1 para indicar espaço vazio
        self.enderecos = np.full(tamanho_max, -1, dtype=int)
        self.count = 0

    def adicionar_endereco(self, endereco):
        """ Adiciona um endereço de página ao bucket se houver espaço disponível. """
        if not self.esta_cheio():
            self.enderecos[self.count] = endereco
            self.count += 1
            return True
        return False  # Bucket cheio

    def esta_cheio(self):
        """ Verifica se o bucket atingiu sua capacidade máxima. """
        return self.count >= self.tamanho_max

    def exibir_conteudo(self):
        """ Exibe os endereços armazenados no bucket. """
        print(f"Bucket -> {list(self.enderecos[:self.count])}")


class Pagina:
    """ Representa uma página de armazenamento contendo registros. """

    def __init__(self, numero, tamanho_max):
        self.numero = numero
        self.tamanho_max = tamanho_max
        self.registros = []  # Lista dinâmica de registros
        self.count = 0

    def adicionar_registro(self, registro):
        if not self.esta_cheia():
            self.registros.append(registro)
            self.count += 1
            return True
        return False  # Página cheia

    def esta_cheia(self):
        return self.count >= self.tamanho_max

    def exibir_registros(self):
        print(f"Página {self.numero}: {self.registros}")

    def buscar_registro(self, indice):
        if 0 <= indice < self.count:
            return self.registros[indice]
        return None


def carregar_dados_arquivo(nome_arquivo, tamanho_pagina):
    """ Carrega os dados do arquivo e divide em páginas de acordo com o tamanho da página. """
    paginas = []
    with open(nome_arquivo, 'r') as file:
        linhas = file.readlines()
        pagina_atual = None
        for i, linha in enumerate(linhas):
            chave = linha.strip()  # Remove o caractere de nova linha

            if pagina_atual is None or pagina_atual.count >= tamanho_pagina:
                pagina_atual = Pagina(i, tamanho_pagina)
                paginas.append(pagina_atual)

            # Adiciona a palavra à página
            pagina_atual.adicionar_registro(chave)
    return paginas


def construir_indice_hash(paginas, indice_hash):
    """ Constrói o índice hash a partir das páginas e suas chaves. """
    cont = 0
    for i, pagina in enumerate(paginas):
        for chave in pagina.registros:
            # Insere a chave e o número da página
            indice_hash.inserir(chave, pagina.numero)
            # print(cont)
            cont += 1


if __name__ == "__main__":
    # Exemplo de uso
    nome_arquivo = 'words.txt'  # Nome do arquivo com os dados
    tamanho_pagina = 5  # Tamanho da página
    num_buckets = 11  # Número de buckets para o índice hash
    tamanho_bucket = 3  # Tamanho máximo de um bucket

    cont = 0

    # Criação do índice hash
    indice_hash = IndiceHash(num_buckets, tamanho_bucket)

    # Carregar dados do arquivo e dividir em páginas
    paginas = carregar_dados_arquivo(nome_arquivo, tamanho_pagina)

    # Construir o índice hash
    construir_indice_hash(paginas, indice_hash)

    # Exibir a tabela do índice
    indice_hash.exibir_tabela()
