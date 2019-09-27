DROP TABLE IF EXISTS starling_transactions;

CREATE TABLE IF NOT EXISTS starling_transactions (
       feed_item_uid UUID PRIMARY KEY,
       category_uid UUID NOT NULL,
       currency TEXT NOT NULL CHECK (currency <> ''),
       minor_units INTEGER NOT NULL,
       source_currency TEXT NOT NULL CHECK (source_currency <> ''),
       source_minor_units INTEGER NOT NULL,
       direction TEXT NOT NULL CHECK (direction <> ''),
       updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
       transaction_time TIMESTAMP WITH TIME ZONE NOT NULL,
       settlement_time TIMESTAMP WITH TIME ZONE,
       source TEXT NOT NULL CHECK (source <> ''),
       source_sub_type TEXT CHECK (source_sub_type <> ''),
       status TEXT NOT NULL CHECK (status <> ''),
       counter_party_type TEXT,
       counter_party_uid UUID,
       counter_party_name TEXT,
       counter_party_sub_entity_uid UUID,
       counter_party_sub_entity_name TEXT,
       counter_party_sub_entity_identifier TEXT CHECK (counter_party_sub_entity_identifier <> ''),
       counter_party_sub_entity_sub_identifier TEXT CHECK (counter_party_sub_entity_sub_identifier <> ''),
       reference TEXT CHECK (reference <> ''),
       country TEXT NOT NULL CHECK (country <> ''),
       spending_category TEXT NOT NULL CHECK (spending_category <> ''),
       user_note TEXT
);
