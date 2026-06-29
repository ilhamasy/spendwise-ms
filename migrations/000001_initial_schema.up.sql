-- 000001_initial_schema.up.sql
-- SpendWise initial database schema

CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    email           VARCHAR(100) NOT NULL,
    password        VARCHAR(255) NOT NULL,
    currency        VARCHAR(10) DEFAULT 'IDR',
    theme           VARCHAR(10) DEFAULT 'system',
    starting_balance BIGINT DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS categories (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36),
    name        VARCHAR(50) NOT NULL,
    type        VARCHAR(10) NOT NULL,
    icon        VARCHAR(10) DEFAULT '',
    color       VARCHAR(10) DEFAULT '',
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);

CREATE TABLE IF NOT EXISTS transactions (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    type        VARCHAR(10) NOT NULL,
    amount      BIGINT NOT NULL,
    category_id VARCHAR(36),
    occurred_at VARCHAR(10) NOT NULL,
    note        VARCHAR(200) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at);

CREATE TABLE IF NOT EXISTS saving_goals (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    target_amount   BIGINT NOT NULL,
    current_saved   BIGINT DEFAULT 0,
    target_date     VARCHAR(10) DEFAULT '',
    status          VARCHAR(10) DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saving_goals_user_id ON saving_goals(user_id);
CREATE INDEX IF NOT EXISTS idx_saving_goals_status ON saving_goals(status);

CREATE TABLE IF NOT EXISTS goal_contributions (
    id          VARCHAR(36) PRIMARY KEY,
    goal_id     VARCHAR(36) NOT NULL,
    amount      BIGINT NOT NULL,
    note        VARCHAR(200) DEFAULT '',
    date        VARCHAR(10) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goal_contributions_goal_id ON goal_contributions(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_contributions_date ON goal_contributions(date);

CREATE TABLE IF NOT EXISTS budgets (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    name        VARCHAR(100) DEFAULT '',
    amount      BIGINT NOT NULL,
    period      VARCHAR(10) NOT NULL,
    category_id VARCHAR(36) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);
CREATE INDEX IF NOT EXISTS idx_budgets_period ON budgets(period);
CREATE INDEX IF NOT EXISTS idx_budgets_category_id ON budgets(category_id);
