<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20260914120000 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Create activity table for authenticated .fit uploads';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('CREATE TABLE activity (id SERIAL NOT NULL, user_id INT NOT NULL, title VARCHAR(180) NOT NULL, type VARCHAR(32) NOT NULL, source VARCHAR(32) NOT NULL, status VARCHAR(32) NOT NULL, started_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL, duration_sec INT NOT NULL, distance_m INT NOT NULL, avg_pace_sec_per_km INT DEFAULT NULL, avg_heart_rate INT DEFAULT NULL, elevation_gain_m INT DEFAULT NULL, summary TEXT DEFAULT NULL, original_filename VARCHAR(255) NOT NULL, stored_path VARCHAR(512) NOT NULL, file_size_bytes INT NOT NULL, created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL, PRIMARY KEY (id))');
        $this->addSql('CREATE INDEX idx_activity_user_started ON activity (user_id, started_at)');
        $this->addSql('ALTER TABLE activity ADD CONSTRAINT FK_AC74095AA76ED395 FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE NOT DEFERRABLE INITIALLY IMMEDIATE');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP CONSTRAINT FK_AC74095AA76ED395');
        $this->addSql('DROP TABLE activity');
    }
}
