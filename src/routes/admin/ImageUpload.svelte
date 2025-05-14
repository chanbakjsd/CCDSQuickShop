<script lang="ts">
	import api from '$lib/api'
	import Button from '$lib/Button.svelte'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'
	import type { ChangeEventHandler } from 'svelte/elements'

	interface Props {
		value: string
	}
	let { value = $bindable() }: Props = $props()

	let fileSelect: HTMLInputElement
	let resolveFileSelect: ((filelist: FileList) => void) | undefined = undefined
	let rejectFileSelect: (() => void) | undefined = undefined
	let uploadError: unknown = $state()
	const upload = async () => {
		fileSelect.click()
		const files = await new Promise<FileList>((resolve, rej) => {
			resolveFileSelect = resolve
			rejectFileSelect = rej
		})
		if (files.length === 0) {
			return
		}
		try {
			value = await api.admin.uploadImage(files[0])
		} catch (e) {
			uploadError = e
		}
	}

	const selectFile: ChangeEventHandler<HTMLInputElement> = (e) => {
		if (!e.currentTarget.files) {
			if (rejectFileSelect) rejectFileSelect()
			return
		}
		if (resolveFileSelect) resolveFileSelect(e.currentTarget.files)
	}
</script>

<div class="flex gap-2">
	<input bind:value />
	<Button onClick={upload} size="md">Upload</Button>
	<ErrorBoundary error={uploadError} />

	<input type="file" onchange={selectFile} bind:this={fileSelect} class="hidden" />
</div>

<style lang="postcss">
	input {
		@apply max-w-lg border border-black px-1;
	}
</style>
