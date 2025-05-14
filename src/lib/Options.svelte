<script lang="ts">
	import { formatPrice } from './cart'
	import type { ShopItemOption } from './shop'

	interface Props {
		options: ShopItemOption[]
		value?: string
		updatePreview?: (preview?: string) => any
	}

	let { options, updatePreview, value = $bindable() }: Props = $props()
	let currentHover: string | undefined = $state()

	const select = (option: string) => () => {
		if (value === option) {
			value = undefined
			return
		}
		value = option
	}
	const hover = (option: string) => () => (currentHover = option)
	const unhover = () => (currentHover = undefined)

	const offsetPriceDisplay = $derived.by(() => {
		const currentOffset = options.find((x) => x.text === value)?.additionalPrice ?? 0
		return options
			.map((x) => (x.additionalPrice ?? 0) - currentOffset)
			.map((x) => {
				if (x > 0) return `(+$${formatPrice(x / 100)})`
				if (x < 0) return `(-$${formatPrice(-x / 100)})`
				return ''
			})
	})

	$effect(() => updatePreview?.(currentHover ?? value))
</script>

<div class="flex flex-wrap gap-x-2 gap-y-1" onpointerleave={unhover}>
	{#each options as option, i}
		<button
			class:selected={value === option.text}
			onclick={select(option.text)}
			onpointerenter={hover(option.text)}
		>
			{option.text}
			{offsetPriceDisplay[i]}
		</button>
	{/each}
</div>

<style lang="postcss">
	button {
		@apply min-w-8 bg-gray-200 px-2 py-1 transition-all hover:bg-gray-300;
		word-break: break-word;
		overflow-break: break-word;
	}
	.selected {
		@apply bg-white hover:bg-white;
		box-shadow: 0 0 1px 2px #0f2b50;
	}
</style>
