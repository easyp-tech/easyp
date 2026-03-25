import Navbar from '../components/Navbar'
import Hero from '../components/Hero'
import TrustedBy from '../components/TrustedBy'
import FeatureSwitcher from '../components/FeatureSwitcher'
import ArchitectureComparison from '../components/ArchitectureComparison'
import Footer from '../components/Footer'

export default function HomePage() {
    return (
        <div className="min-h-screen bg-background text-white">
            <Navbar />
            <Hero />
            <TrustedBy />
            <FeatureSwitcher />
            <ArchitectureComparison />
            <Footer />
        </div>
    )
}
