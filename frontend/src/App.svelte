<script>
  // Add RepairAndSave to your imports
  import { AnalyzeFile, SelectFile, GetBitPlaneImages, GetForensicFilters, RepairAndSave } from '../wailsjs/go/main/App.js'

  let filePath = ""
  let results = null
  let bitPlaneMaps = []
  let forensicFilters = []
  let errorMessage = ""
  let loading = false
  
  // NEW: State for the repair logic
  let hasStructureIssue = false

  async function handleBrowse() {
    const selection = await SelectFile()
    if (selection) {
      filePath = selection
      await startAnalysis()
    }
  }
  async function startAnalysis() {
    if (!filePath) return
    try {
        loading = true
        errorMessage = ""
        hasStructureIssue = false 
        
        // 1. Get findings first (this usually succeeds even if the image is "broken")
        results = await AnalyzeFile(filePath)

        // 2. Check for structural issues immediately
        if (results && results.findings) {
            hasStructureIssue = results.findings.some(f => 
                f.analyzer_name === "Header & Structure Analyzer" && 
                (f.confidence === "Critical" || f.confidence === "Medium")
            );
        }

        // 3. Try to get images (this is what throws the "unexpected EOF")
        bitPlaneMaps = await GetBitPlaneImages(filePath)
        forensicFilters = await GetForensicFilters(filePath)

    } catch (err) {
        // Even if images fail, we keep the results so the repair button can show
        errorMessage = "Note: Image decoding failed (" + err + "). Structure repair may be required."
    } finally {
        loading = false
    }
  }

  // NEW: Function to trigger the Go repair logic
  async function handleRepair() {
    try {
      loading = true
      const newPath = await RepairAndSave(filePath)
      alert("Header repaired successfully! Saved as: " + newPath)
      
      // Automatically switch to the repaired file and re-scan
      filePath = newPath
      await startAnalysis()
    } catch (err) {
      errorMessage = "Repair failed: " + err
    } finally {
      loading = false
    }
  }
</script>

<main>
  <header>
    <h1>🔍 InPlainSight</h1>
    <p>Forensic Bit-Plane Analysis Dashboard</p>
  </header>

  <div class="input-box">
    <input bind:value={filePath} placeholder="Click Browse to select an image..." class="input" readonly />
    <button class="btn" on:click={handleBrowse} disabled={loading}>
      {loading ? "Analyzing..." : "Browse & Scan"}
    </button>
  </div>

  {#if errorMessage}
    <div class="error">{errorMessage}</div>
  {/if}

  {#if hasStructureIssue}
    <div class="repair-banner">
      <span>⚠️ Structural inconsistencies detected in file header.</span>
      <button class="repair-btn" on:click={handleRepair} disabled={loading}>
        {loading ? "Repairing..." : "Fix Header Markers"}
      </button>
    </div>
  {/if}

  {#if forensicFilters.length > 0}
    <section class="visualizer">
      <h3>Forensic Enhancement Filters</h3>
      <div class="map-grid" style="grid-template-columns: repeat(2, 1fr);">
        {#each forensicFilters as filter}
          <div class="map-card">
            <img src={filter.data} alt={filter.name} />
            <span>{filter.name}</span>
          </div>
        {/each}
      </div>
      <p class="hint">Inversion reveals dark-channel stego; Contrast Stretching exposes "Ghost" artifacts and clone-stamp traces.</p>
    </section>
  {/if}

  {#if results}
    <section class="visualizer">
      <h3>Bit-Plane Noise Maps (0-3)</h3>
      <div class="map-grid">
        {#each bitPlaneMaps as map, i}
          <div class="map-card">
            <img src={map} alt="Bit Plane {i}" />
            <span>Plane {i}</span>
          </div>
        {/each}
      </div>
      <p class="hint">Look for "Static" patterns. Clean images show ghostly outlines; stego-images show random noise.</p>
    </section>

    <div class="findings-grid">
      {#each results.findings as finding}
        <div class="card {finding.confidence.toLowerCase()}">
          <div class="card-header">
            <strong>{finding.analyzer_name}</strong>
            <span class="badge">{finding.confidence}</span>
          </div>
          <p>{finding.description}</p>
          {#if finding.data_found}
            <div class="code-box"><code>{finding.data_found}</code></div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</main>

<style>
  :global(body) { background: #0f172a; color: #f8fafc; font-family: sans-serif; margin: 0; }
  main { padding: 2rem; max-width: 1100px; margin: 0 auto; }
  header { text-align: center; margin-bottom: 2rem; border-bottom: 1px solid #334155; padding-bottom: 1rem; }
  h1 { color: #38bdf8; margin: 0; }

  .input-box { display: flex; gap: 1rem; margin-bottom: 2rem; }
  .input { flex: 1; padding: 0.8rem; border-radius: 6px; border: 1px solid #334155; background: #1e293b; color: #94a3b8; }
  .btn { padding: 0.8rem 2rem; background: #38bdf8; color: #0f172a; border: none; border-radius: 6px; font-weight: bold; cursor: pointer; }

  /* NEW: Repair Banner Styles */
  .repair-banner {
    background: #451a03;
    border: 1px solid #f59e0b;
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: #fef3c7;
  }
  .repair-btn {
    background: #f59e0b;
    color: #451a03;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: bold;
    cursor: pointer;
  }
  .repair-btn:hover { background: #fbbf24; }

  .visualizer { background: #1e293b; padding: 1.5rem; border-radius: 12px; margin-bottom: 2rem; }
  .map-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-top: 1rem; }
  .map-card { text-align: center; background: #0f172a; padding: 0.5rem; border-radius: 6px; }
  .map-card img { width: 100%; height: auto; border-radius: 4px; image-rendering: pixelated; }
  .map-card span { font-size: 0.7rem; color: #64748b; text-transform: uppercase; }
  .hint { font-size: 0.8rem; color: #94a3b8; margin-top: 1rem; font-style: italic; }

  .findings-grid { display: grid; gap: 1rem; }
  .card { background: #1e293b; padding: 1.2rem; border-radius: 8px; border-left: 4px solid #475569; }
  .card.critical { border-left-color: #ef4444; }
  .card.high { border-left-color: #f97316; }
  .card-header { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
  .badge { font-size: 0.7rem; background: #334155; padding: 2px 8px; border-radius: 4px; }
  .code-box { background: #0f172a; padding: 0.8rem; border-radius: 4px; margin-top: 1rem; }
  code { color: #10b981; font-family: monospace; font-size: 0.85rem; word-break: break-all; }
  .error { background: #450a0a; color: #fecaca; padding: 1rem; border-radius: 6px; margin-bottom: 1rem; }
</style>
