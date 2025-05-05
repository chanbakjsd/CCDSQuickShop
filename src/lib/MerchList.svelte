<script lang="ts">
	import MerchCard from './MerchCard.svelte'
	import { resolveImageURL, type ShopItem } from './shop'

	interface Props {
		items: ShopItem[]
		variants: Record<string, string>[]
		value: number
	}

	let { items, variants, value = $bindable(-1) }: Props = $props()

	const select = (i: number) => () => {
		if (value === i) {
			value = -1
			return
		}
		value = i
	}

	const previewImages = $derived(items.map((x, i) => resolveImageURL(x, variants[i])))
</script>

<div class="flex flex-wrap gap-4">
	{#each items as item, i}
		<MerchCard
			name={item.name}
			imageURL={previewImages[i]}
			basePrice={item.basePrice / 100}
			selected={value === i}
			on:click={select(i)}
		/>
	{/each}
</div>
