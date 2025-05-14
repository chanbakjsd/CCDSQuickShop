<script lang="ts">
	import type { Snippet } from 'svelte'
	import Icon from './icon/Icon.svelte'

	interface Props {
		children: Snippet
		onClick: () => any
		disabled?: boolean
		size?: 'md' | 'lg'
		loading?: boolean
	}
	const {
		children,
		onClick,
		disabled = false,
		size = 'lg',
		loading: forceLoading = false
	}: Props = $props()

	let onClickPending = $state(false)
	export const click = async () => {
		if (shouldDisable) return
		try {
			onClickPending = true
			await onClick()
		} catch (e) {
			console.error('Unhandled onClick error', e)
		} finally {
			onClickPending = false
		}
	}

	const isLoading = $derived(forceLoading || onClickPending)
	const shouldDisable = $derived(disabled || isLoading)
	const contentClass = $derived(isLoading ? 'opacity-0' : 'opacity-100')
	const loaderClass = $derived(isLoading ? 'opacity-100' : 'opacity-0')
</script>

<button class={`relative size-${size}`} onclick={click} disabled={shouldDisable}>
	<div class={`transition-all ${contentClass}`}>{@render children()}</div>
	<div class={`absolute bottom-1/3 left-1/3 right-1/3 top-1/3 flex ${loaderClass}`}>
		<div class="mx-auto"><Icon name="loading" /></div>
	</div>
</button>

<style lang="postcss">
	button {
		@apply rounded-full bg-brand px-6 py-2 text-lg text-white transition-all;
		@apply flex items-center justify-center gap-2;
	}
	button:hover {
		@apply scale-105;
	}
	button:active {
		@apply scale-100;
	}
	button:disabled {
		@apply scale-100 cursor-not-allowed bg-gray-300 text-gray-700;
	}

	button.size-md {
		@apply px-4 py-1 text-base;
	}
</style>
