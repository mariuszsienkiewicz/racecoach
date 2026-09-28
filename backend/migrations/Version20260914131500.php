<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20260914131500 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add MinIO/S3 object key fields to activity and make stored_path nullable';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD object_key VARCHAR(512) DEFAULT NULL');
        $this->addSql('ALTER TABLE activity ADD storage_bucket VARCHAR(128) DEFAULT NULL');
        $this->addSql('ALTER TABLE activity ALTER stored_path DROP NOT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ALTER stored_path SET NOT NULL');
        $this->addSql('ALTER TABLE activity DROP object_key');
        $this->addSql('ALTER TABLE activity DROP storage_bucket');
    }
}
