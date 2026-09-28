<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20260915130134 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add activity.features_object_key for AI feature payload artifact';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD features_object_key VARCHAR(512) DEFAULT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP features_object_key');
    }
}
