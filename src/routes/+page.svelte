<script lang="ts">
	import type { CartItem } from '$lib/cart';
	import Header from '$lib/Header.svelte';

	import type { PageData } from './$types';
	import Cart from './Cart.svelte';
	import MerchSelection from './MerchSelection.svelte';

	const { data }: { data: PageData } = $props();
	let cart: CartItem[] = $state([]);

	const addToCart = (item: CartItem) => {
		// Try to search for existing entries and add amount instead.
		for (let i = 0; i < cart.length; i++) {
			if (
				cart[i].id !== item.id ||
				cart[i].variant.length !== item.variant.length ||
				cart[i].amount + item.amount > 50
			) {
				continue;
			}
			const variantMatch = cart[i].variant.every((x) =>
				item.variant.some((y) => x.type === y.type && x.option === y.option)
			);
			if (!variantMatch) {
				continue;
			}
			// This is a match! Just add amount.
			cart[i].amount += item.amount;
			return;
		}
		cart = [...cart, item];
	};
</script>

<svelte:head><title>SCDS Merch Store</title></svelte:head>

<main class="grid grid-cols-1 md:grid-cols-3">
	<div class="left-panel">
		<Header />
		<div class="px-2 md:py-2">
			<MerchSelection items={data.items} addItem={addToCart} />
		</div>
	</div>
	<div class="right-panel"><Cart bind:cart availableCoupons={data.coupons} /></div>
</main>

<style lang="postcss">
	.left-panel {
		@apply col-span-2 mx-2 h-fit border-b border-gray-400 pb-4 pt-2;
		@apply flex flex-col gap-4;
		@apply md:mx-0 md:my-2 md:border-b-0 md:border-r md:px-4 md:py-0;
		@apply md:h-[calc(100vh-1rem)] md:overflow-y-auto;
	}
	.right-panel {
		@apply overflow-y-auto px-4 py-4;
		@apply md:h-screen;
	}
</style>
