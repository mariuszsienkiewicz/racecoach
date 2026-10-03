<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

final class Version20261003123930 extends AbstractMigration
{
    public function getDescription(): string
    {
        return 'Add activity.analysis_version for pipeline reprocess eligibility';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity ADD analysis_version INT DEFAULT NULL');
    }

    public function down(Schema $schema): void
    {
        $this->addSql('ALTER TABLE activity DROP analysis_version');
    }
}
