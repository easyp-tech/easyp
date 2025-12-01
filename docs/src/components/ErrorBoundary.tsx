import { Component, ErrorInfo, ReactNode } from 'react'

interface Props {
    children: ReactNode
}

interface State {
    hasError: boolean
    error: Error | null
}

export default class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null
    }

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error }
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('Uncaught error:', error, errorInfo)
    }

    public render() {
        if (this.state.hasError) {
            return (
                <div className="p-6 bg-red-500/10 border border-red-500/20 rounded-lg text-red-200">
                    <h2 className="text-xl font-bold mb-2">Something went wrong</h2>
                    <p className="mb-4">We encountered an error while rendering this page.</p>
                    <pre className="bg-slate-900 p-4 rounded overflow-auto text-sm font-mono">
                        {this.state.error?.toString()}
                    </pre>
                </div>
            )
        }

        return this.props.children
    }
}
