<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20260914160000 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add metrics_object_key to activity for worker-produced metrics.json';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD metrics_object_key VARCHAR(512) DEFAULT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP metrics_object_key');
    }
}
