<script>
  import { AnalyzeFile, SelectFile, GetBitPlaneImages, GetForensicFilters, GetAudioSpectrogram, RepairAndSave } from '../wailsjs/go/main/App.js'

  let filePath = ""
  let results = null
  let bitPlaneMaps = []
  let forensicFilters = []
  let errorMessage = ""
  let loading = false
  let hasStructureIssue = false
  let spectrogram = ""
  
  // NEW: State to distinguish between Image and Audio UI
  let isAudio = false

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
        bitPlaneMaps = []
        forensicFilters = []
        spectrogram = ""

        // 1. Run the core Analysis
        results = await AnalyzeFile(filePath)
        
        // Check file type immediately
        
        if (results && results.findings) {
            hasStructureIssue = results.findings.some(f =>
                (f.analyzer_name === "Header & Structure Analyzer" || f.analyzer_name === "ID3 Tag Investigator" || f.analyzer_name === "Audio LSB Bit-Plane Analyzer") &&
                (f.confidence === "Critical" || f.confidence === "Medium" || f.confidence === "High")
            );
        }

        isAudio = results.file_type.toLowerCase().includes("audio")

        if (isAudio) {
            // Fetch the spectrogram for audio files
            spectrogram = await GetAudioSpectrogram(filePath)
        } else {
            // Fetch maps for images
            bitPlaneMaps = await GetBitPlaneImages(filePath)
            forensicFilters = await GetForensicFilters(filePath)
        }
        
        // 3. IMAGE ONLY: Run visual analysis if it's not audio
        if (!isAudio) {
            try {
                bitPlaneMaps = await GetBitPlaneImages(filePath)
                forensicFilters = await GetForensicFilters(filePath)
            } catch (imageErr) {
                errorMessage = "Visual decoding failed. Structural repair may be required for image preview."
            }
        }

    } catch (err) {
        errorMessage = "Analysis Error: " + err
    } finally {
        loading = false
    }
  }

  async function handleRepair() {
    try {
      loading = true
      const newPath = await RepairAndSave(filePath)
      alert("Evidence sanitized successfully! Saved as: " + newPath)
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
    <p>Multi-Modal Forensic Analysis Dashboard</p>
  </header>

  <div class="input-box">
    <input bind:value={filePath} placeholder="Select image or audio evidence..." class="input" readonly />
    <button class="btn" on:click={handleBrowse} disabled={loading}>
      {loading ? "Analyzing..." : "Browse & Scan"}
    </button>
  </div>

  {#if errorMessage}
    <div class="error">{errorMessage}</div>
  {/if}

  {#if hasStructureIssue}
    <div class="repair-banner">
      <span>⚠️ Anomalies detected in file structure/metadata.</span>
      <button class="repair-btn" on:click={handleRepair} disabled={loading}>
        {loading ? "Repairing..." : "Sanitize File"}
      </button>
    </div>
  {/if}

  {#if isAudio && results}
  <div class="audio-container">
    <div class="audio-hero">
      <div class="audio-icon">🔊</div>
      <div class="audio-info">
          <h3>{results.file_type} Forensic Profile</h3>
          <p>Signal-to-Noise Floor and Frequency Domain Analysis active.</p>
      </div>
    </div>

    {#if spectrogram}
      <section class="visualizer">
        <h3>Frequency Spectrogram (FFT Heatmap)</h3>
        <div class="spectro-frame">
          <img src={spectrogram} alt="Frequency Spectrogram" class="spectro-img" />
        </div>
        <p class="hint">Frequency (Y-axis) vs Time (X-axis). Intense color bands indicate coherent signals; horizontal static indicates noise-floor manipulation.</p>
      </section>
    {/if}
  </div>
  {/if}

  {#if !isAudio && forensicFilters.length > 0}
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
    </section>
  {/if}

  {#if !isAudio && bitPlaneMaps.length > 0}
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
    </section>
  {/if}

  {#if results}
    <div class="findings-grid">
      <h3>Analysis Findings</h3>
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
  /* Keep your existing styles and add these new ones */
  :global(body) { background: #0f172a; color: #f8fafc; font-family: sans-serif; margin: 0; }
  main { padding: 2rem; max-width: 1100px; margin: 0 auto; }
  
  /* ... [Your previous styles for .card, .btn, .map-grid, etc.] ... */

  .audio-card {
      background: linear-gradient(90deg, #1e293b 0%, #0f172a 100%);
      border: 1px solid #38bdf8;
      padding: 2rem;
      border-radius: 12px;
      display: flex;
      align-items: center;
      gap: 2rem;
      margin-bottom: 2rem;
  }
  .audio-icon { font-size: 3rem; }
  .audio-info h3 { margin: 0; color: #38bdf8; }
  .audio-info p { margin: 0.5rem 0 0 0; color: #94a3b8; }

  .audio-hero {
    background: linear-gradient(90deg, #1e293b 0%, #0f172a 100%);
    border: 1px solid #38bdf8;
    padding: 1.5rem;
    border-radius: 12px;
    display: flex;
    align-items: center;
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .spectro-frame {
    background: #000;
    padding: 10px;
    border-radius: 8px;
    border: 1px solid #334155;
    overflow-x: auto; /* Allow scrolling if the audio is long */
  }

  .spectro-img {
    width: 100%;
    height: 300px; /* Fixed height looks better for spectrograms */
    object-fit: fill;
    display: block;
    image-rendering: pixelated; /* Keeps the FFT bins sharp */
  }

  /* Re-add your previous styles here to ensure buttons and inputs stay styled */
  .input-box { display: flex; gap: 1rem; margin-bottom: 2rem; }
  .input { flex: 1; padding: 0.8rem; border-radius: 6px; border: 1px solid #334155; background: #1e293b; color: #94a3b8; }
  .btn { padding: 0.8rem 2rem; background: #38bdf8; color: #0f172a; border: none; border-radius: 6px; font-weight: bold; cursor: pointer; }
  .repair-banner { background: #451a03; border: 1px solid #f59e0b; padding: 1rem; border-radius: 8px; margin-bottom: 2rem; display: flex; justify-content: space-between; align-items: center; color: #fef3c7; }
  .repair-btn { background: #f59e0b; color: #451a03; border: none; padding: 0.5rem 1rem; border-radius: 4px; font-weight: bold; cursor: pointer; }
  .visualizer { background: #1e293b; padding: 1.5rem; border-radius: 12px; margin-bottom: 2rem; }
  .map-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-top: 1rem; }
  .map-card { text-align: center; background: #0f172a; padding: 0.5rem; border-radius: 6px; }
  .map-card img { width: 100%; height: auto; border-radius: 4px; image-rendering: pixelated; }
  .map-card span { font-size: 0.7rem; color: #64748b; text-transform: uppercase; }
  .findings-grid { display: grid; gap: 1rem; }
  .card { background: #1e293b; padding: 1.2rem; border-radius: 8px; border-left: 4px solid #475569; }
  .card.critical { border-left-color: #ef4444; }
  .card.high { border-left-color: #f97316; }
  .card.medium { border-left-color: #eab308; }
  .card-header { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
  .badge { font-size: 0.7rem; background: #334155; padding: 2px 8px; border-radius: 4px; }
  .code-box { background: #0f172a; padding: 0.8rem; border-radius: 4px; margin-top: 1rem; }
  code { color: #10b981; font-family: monospace; font-size: 0.85rem; word-break: break-all; }
  .error { background: #450a0a; color: #fecaca; padding: 1rem; border-radius: 6px; margin-bottom: 1rem; }
</style>
