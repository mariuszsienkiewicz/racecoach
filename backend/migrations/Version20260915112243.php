<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20260915150000 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add per-user FIT checksum for duplicate upload rejection';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD checksum_sha256 VARCHAR(64) DEFAULT NULL');
        $this->addSql('CREATE UNIQUE INDEX uniq_activity_user_checksum ON activity (user_id, checksum_sha256) WHERE checksum_sha256 IS NOT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('DROP INDEX uniq_activity_user_checksum');
        $this->addSql('ALTER TABLE activity DROP checksum_sha256');
    }
}
