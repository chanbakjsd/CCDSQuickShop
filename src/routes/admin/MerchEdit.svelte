<script lang="ts">
	import { onMount } from 'svelte';

	import { fetchProducts, updateProduct } from '$lib/api';
	import { emptyShopItem, type ShopItem } from '$lib/shop';
	import MerchList from '$lib/MerchList.svelte';
	import ProductEdit from './ProductEdit.svelte';

	let loading = true;
	let items: ShopItem[] = [];
	onMount(() => {
		fetchProducts(true).then((x) => {
			loading = false;
			items = x;
		});
	});

	$: {
		if (!loading && (items.length === 0 || items[items.length - 1].id !== '')) {
			items.push(emptyShopItem('Add New Product'));
		}
	}

	$: variants = Array(items.length).fill({});
	let selectedItemIdx = -1;
	$: selectedItem = selectedItemIdx >= 0 ? items[selectedItemIdx] : undefined;

	const update = async () => {
		if (!selectedItem) return;
		const updatedItem = await updateProduct(selectedItem);
		items[selectedItemIdx] = updatedItem;
	};
</script>

<div class="flex flex-col gap-4 p-4">
	<MerchList {items} {variants} bind:value={selectedItemIdx} />
	{#if selectedItem}
		<hr />
		<ProductEdit bind:product={selectedItem} updateProduct={update} />
	{/if}
</div>
