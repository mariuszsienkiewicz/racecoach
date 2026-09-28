<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20260915123359 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('DROP INDEX uniq_activity_user_checksum');
        $this->addSql('ALTER TABLE activity ADD structure_object_key VARCHAR(512) DEFAULT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP structure_object_key');
        $this->addSql('CREATE UNIQUE INDEX uniq_activity_user_checksum ON activity (user_id, checksum_sha256) WHERE (checksum_sha256 IS NOT NULL)');
    }
}
