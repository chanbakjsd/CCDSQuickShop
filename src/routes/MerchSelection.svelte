<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { formatPrice, type CartItem } from '$lib/cart';
	import { resolveImageURL, tentativePrice, toArrayVariant, type ShopItem } from '$lib/shop';
	import Options from '$lib/Options.svelte';
	import Button from '$lib/Button.svelte';
	import MerchList from '$lib/MerchList.svelte';

	export let items: ShopItem[];
	export let addItem: (item: CartItem) => void;

	// All the variants that the user has committed to (clicked on).
	const chosenVariants: Record<string, string>[] = [];
	$: {
		while (chosenVariants.length < items.length) chosenVariants.push({});
	}
	// The index of the merch being selected right now.
	let activeMerch = -1;
	$: selectedItem = activeMerch >= 0 ? items[activeMerch] : undefined;

	// The cart item representing the user's current choice.
	let cartItem: CartItem | undefined;
	$: {
		if (!selectedItem) {
			cartItem = undefined;
			break $;
		}
		const activeVariant = chosenVariants[activeMerch];
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

	// The variant being previewed (selected), gets updated due to bind below.
	$: activePreviewVariant = activeMerch >= 0 ? chosenVariants[activeMerch] : {};
	// The chosenVariants table but replaces the selected option with the preview variant to allow sneak peeks.
	$: previewVariants = items.map((_, i) =>
		i === activeMerch ? activePreviewVariant : chosenVariants[i] || {}
	);
	const tryAddItem = () => {
		if (!cartItem) return;
		addItem({ ...cartItem });
	};
</script>

<div class="flex flex-col gap-4">
	<MerchList {items} variants={previewVariants} bind:value={activeMerch} />
	{#if selectedItem}
		<hr class="border-gray-300" transition:fade />
		<div class="flex flex-col gap-2" transition:fly={{ y: 100 }}>
			<p class="text-3xl">{selectedItem.name}</p>
			<p class="text-xl">
				S$ {formatPrice(tentativePrice(selectedItem, activePreviewVariant) / 100)}
			</p>
			<div class="options">
				{#each selectedItem.variants as variant}
					<span class="py-1 transition-all">{variant.type}</span>
					<Options
						options={variant.options}
						bind:value={chosenVariants[activeMerch][variant.type]}
						bind:previewValue={activePreviewVariant[variant.type]}
					/>
				{/each}
			</div>
			<div class="mt-2">
				<Button onClick={tryAddItem} disabled={!cartItem}>Add to Cart</Button>
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
