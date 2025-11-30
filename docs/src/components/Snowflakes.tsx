import { useEffect, useState } from 'react'

interface Snowflake {
    id: number
    x: number
    size: number
    animationDuration: number
    animationDelay: number
    symbol: string
}

export default function Snowflakes() {
    const [snowflakes, setSnowflakes] = useState<Snowflake[]>([])

    useEffect(() => {
        const snowflakeSymbols = ['❄', '❅', '❆']
        const flakes: Snowflake[] = []

        // Create 16 snowflakes
        for (let i = 0; i < 16; i++) {
            flakes.push({
                id: i,
                x: Math.random() * 100, // Random horizontal position (0-100%)
                size: Math.random() * 0.8 + 0.5, // Random size (0.5-1.3)
                animationDuration: Math.random() * 8 + 8, // Random duration (8-16s)
                animationDelay: Math.random() * 16 - 8, // Random delay from -8s to +8s
                symbol: snowflakeSymbols[Math.floor(Math.random() * snowflakeSymbols.length)]
            })
        }

        setSnowflakes(flakes)
    }, [])

    return (
        <div className="fixed inset-0 pointer-events-none z-10 overflow-hidden">
            {snowflakes.map((flake) => (
                <div
                    key={flake.id}
                    className="absolute text-white/40 select-none"
                    style={{
                        left: `${flake.x}%`,
                        fontSize: `${flake.size}rem`,
                        animationName: 'snowfall',
                        animationDuration: `${flake.animationDuration}s`,
                        animationDelay: `${flake.animationDelay}s`,
                        animationIterationCount: 'infinite',
                        animationTimingFunction: 'linear'
                    }}
                >
                    {flake.symbol}
                </div>
            ))}
        </div>
    )
}
