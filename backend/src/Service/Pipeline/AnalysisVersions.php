<?php

namespace App\Service\Pipeline;

final class AnalysisVersions
{
    public const CURRENT = 2;

    public const REPROCESS_RECENT_LIMIT = 5;

    public const STALE_PIPELINE_AFTER = '3 minutes';
}
