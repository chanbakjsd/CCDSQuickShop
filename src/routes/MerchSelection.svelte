<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { formatPrice, type CartItem } from '$lib/cart';
	import { resolveImageURL, tentativePrice, toArrayVariant, type ShopItem } from '$lib/shop';
	import Options from '$lib/Options.svelte';
	import Card from './Card.svelte';
	import Button from '$lib/Button.svelte';

	export let items: ShopItem[];
	export let addItem: (item: CartItem) => void;

	type SelectedVariants = Record<string, string>;
	const chosenVariants: Record<string, SelectedVariants> = {};
	let activeVariant: SelectedVariants = {};
	let activePreviewVariant: SelectedVariants = {};
	let selectedItemIdx = -1;
	const select = (i: number) => () => {
		// Save previous variant selections if applicable.
		if (selectedItemIdx !== -1) {
			const item = items[selectedItemIdx];
			chosenVariants[item.id] = activeVariant;
		}
		// Unselect.
		if (selectedItemIdx === i) {
			selectedItemIdx = -1;
			activeVariant = {};
			activePreviewVariant = {};
			return;
		}
		// Select.
		selectedItemIdx = i;
		const item = items[i];
		if (item.id in chosenVariants) {
			activeVariant = chosenVariants[item.id];
		} else {
			activeVariant = {};
		}
		activePreviewVariant = activeVariant;
	};
	$: selectedItem = selectedItemIdx >= 0 ? items[selectedItemIdx] : undefined;
	$: previewImages = items.map((x, i) => {
		const previewVariant =
			i === selectedItemIdx ? activePreviewVariant : chosenVariants[x.id] || {};
		return resolveImageURL(x, previewVariant);
	});

	let cartItem: CartItem | undefined;
	$: {
		if (!selectedItem) {
			cartItem = undefined;
			break $;
		}
		const arrayVariant = toArrayVariant(selectedItem, activeVariant);
		// Everything must be selected.
		if (arrayVariant.some((x) => !x)) {
			cartItem = undefined;
			break $;
		}
		cartItem = {
			id: selectedItem.id,
			name: selectedItem.name,
			variant: arrayVariant.map((x, i) => ({ type: selectedItem.variants[i].type, option: x! })),
			imageURL: resolveImageURL(selectedItem, activeVariant),
			unitPrice: tentativePrice(selectedItem, activeVariant),
			amount: 1
		};
	}

	const tryAddItem = () => {
		if (!cartItem) return;
		addItem({ ...cartItem });
	};
</script>

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap gap-4">
		{#each items as item, i}
			<Card
				name={item.name}
				imageURL={previewImages[i]}
				basePrice={item.basePrice / 100}
				selected={selectedItemIdx === i}
				on:click={select(i)}
			/>
		{/each}
	</div>
	{#if selectedItem}
		<hr class="border-gray-300" transition:fade />
		<div class="flex flex-col gap-2" transition:fly={{ y: 100 }}>
			<p class="text-3xl">{selectedItem.name}</p>
			<p class="text-xl">S$ {formatPrice(tentativePrice(selectedItem, activeVariant) / 100)}</p>
			<div class="options">
				{#each selectedItem.variants as variant}
					<span class="py-1 transition-all">{variant.type}</span>
					<Options
						options={variant.options}
						bind:value={activeVariant[variant.type]}
						bind:previewValue={activePreviewVariant[variant.type]}
					/>
				{/each}
			</div>
			<div class="mt-2">
				<Button on:click={tryAddItem} disabled={!cartItem}>Add to Cart</Button>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	.options {
		@apply grid gap-x-4 gap-y-2;
		grid-template-columns: auto 1fr;
	}
</style>
