<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20261003132835 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add activity.updated_at for stale pipeline reprocess UX';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD updated_at TIMESTAMP(0) WITHOUT TIME ZONE DEFAULT NULL');
        $this->addSql('UPDATE activity SET updated_at = created_at WHERE updated_at IS NULL');
        $this->addSql('ALTER TABLE activity ALTER COLUMN updated_at SET NOT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP updated_at');
    }
}
