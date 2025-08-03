<script lang="ts">
	import MerchCard from './MerchCard.svelte'
	import { resolveImageURL, type ShopItem } from './shop'

	interface Props {
		items: ShopItem[]
		variants: Record<string, string | undefined>[]
		value: number
	}

	let { items, variants, value = $bindable(-1) }: Props = $props()
	const select = (i: number) => () => (value = value === i ? -1 : i)
</script>

<div class="flex flex-wrap gap-4">
	{#each items as item, i}
		<MerchCard
			name={item.name}
			imageURL={resolveImageURL(item, variants[i] || {})}
			basePrice={item.basePrice / 100}
			selected={value === i}
			onclick={select(i)}
		/>
	{/each}
</div>
