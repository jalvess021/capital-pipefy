CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    tipo_solicitacao VARCHAR(100) NOT NULL,
    valor_patrimonio DECIMAL(15,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Aguardando Análise',
    prioridade VARCHAR(50) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_email ON clients(email);
