import { FinalCtaSection, FlowSection } from '@/components/landing/flow-section'
import { AnalysisSection } from '@/components/landing/analysis-section'
import { CoachSection } from '@/components/landing/coach-section'
import { LandingHero } from '@/components/landing/landing-hero'
import { PlansSection } from '@/components/landing/plans-section'
import { SyncSection } from '@/components/landing/sync-section'
import { SiteFooter } from '@/components/layout/site-footer'
import { SiteHeader } from '@/components/layout/site-header'

export function HomePage() {
  return (
    <div id="top" className="relative min-h-svh">
      <SiteHeader />
      <main>
        <LandingHero />
        <SyncSection />
        <AnalysisSection />
        <PlansSection />
        <CoachSection />
        <FlowSection />
        <FinalCtaSection />
      </main>
      <SiteFooter />
    </div>
  )
}
