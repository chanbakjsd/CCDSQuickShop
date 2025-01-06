<script lang="ts">
	import MerchCard from './MerchCard.svelte';
	import { resolveImageURL, type ShopItem } from './shop';

	export let items: ShopItem[];
	export let variants: Record<string, string>[];
	export let value = -1;

	const select = (i: number) => () => {
		if (value === i) {
			value = -1;
			return;
		}
		value = i;
	};

	$: previewImages = items.map((x, i) => resolveImageURL(x, variants[i]));
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
