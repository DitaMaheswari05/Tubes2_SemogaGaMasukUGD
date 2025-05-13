"use client";

import { useState, useEffect, useRef } from "react";
import SearchTreeView from "./SearchTreeView";

function LiveSearchVisualizer({ searchSteps, targetElement }) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(500); // ms
  const timerRef = useRef(null);

  // Debug logging - add this
  useEffect(() => {
    console.log("Search steps received:", searchSteps?.length);
    console.log("First step sample:", searchSteps?.[0]);
  }, [searchSteps]);

  // Convert discovered edges to a recipe format for visualization
  const buildDiscoveredRecipes = (discoveredEdges) => {
    if (!discoveredEdges) {
      console.log("No discovered edges data");
      return {};
    }

    console.log("Building from edges:", discoveredEdges);

    const recipes = {};
    // Handle both possible data formats
    Object.entries(discoveredEdges).forEach(([product, info]) => {
      // Check if we're dealing with the ID format or name format
      if (info && typeof info === "object") {
        recipes[product] = {
          A: info.A || (info.ParentID ? targetElement : "Unknown"),
          B: info.B || (info.PartnerID ? targetElement : "Unknown"),
        };
      }
    });
    return recipes;
  };

  // Ensure we have data
  if (!searchSteps || searchSteps.length === 0) {
    return <div>No search data available. Make sure BFS/DFS returned search steps.</div>;
  }

  // Auto-play animation
  useEffect(() => {
    if (isPlaying && searchSteps) {
      timerRef.current = setInterval(() => {
        setCurrentStep((prev) => {
          if (prev >= searchSteps.length - 1) {
            setIsPlaying(false);
            return prev;
          }
          const nextStep = searchSteps[prev + 1];
          const foundTarget = nextStep?.foundTarget;

          // If we found the target, make sure to show that step
          if (foundTarget) {
            setIsPlaying(false); // Auto-pause when target is found
          }
          return prev + 1;
        });
      }, playbackSpeed);
    }

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [isPlaying, searchSteps, playbackSpeed]);

  // Reset when searchSteps changes
  useEffect(() => {
    setCurrentStep(0);
    setIsPlaying(false);
  }, [searchSteps]);

  if (!searchSteps || !searchSteps.length) {
    return <div>No search data available</div>;
  }

  const step = searchSteps[currentStep];
  console.log("Current step details:", {
    stepNumber: currentStep,
    stepData: step,
    hasDiscovered: !!step.discovered,
    discoveredKeys: step.discovered ? Object.keys(step.discovered) : [],
    currentNode: step.current,
    queueSize: step.queue?.length,
  });
  const discovered = buildDiscoveredRecipes(step.discovered || {});

  return (
    <div className="live-search-visualizer">
      <h3>Search Visualization for "{targetElement}"</h3>

      <div
        className="step-info"
        style={{
          backgroundColor: "#f5f5f5",
          padding: "10px",
          borderRadius: "5px",
          marginBottom: "20px",
        }}>
        <div>
          <strong>Step:</strong> {step.step} of {searchSteps.length - 1}
        </div>
        <div>
          <strong>Current Element:</strong> {step.current || "None"}
        </div>
        <div>
          <strong>Queue Size:</strong> {step.queue?.length || 0} elements
        </div>
        <div>
          <strong>Queue Elements: </strong> {(step.queue || []).join(", ") || "Empty"}
        </div>
        <div>
          <strong>Discovered Objects </strong> {Object.keys(discovered).length}:
        </div>
        <div style={{ maxHeight: "200px", overflow: "auto", fontSize: "12px" }}>
          <pre style={{ backgroundColor: "#f5f5f5", padding: "10px" }}>
            {JSON.stringify(discovered, null, 2) || "No discoveries"}
          </pre>
        </div>
        <div>
          <strong>Target Found:</strong> {step.found_target === true ? "Yes! 🎉" : "Still searching..."}
        </div>
      </div>
      <div
        className="progress-bar"
        style={{
          width: "100%",
          height: "10px",
          backgroundColor: "#eee",
          marginBottom: "10px",
          position: "relative",
        }}>
        <div
          style={{
            width: `${(currentStep / (searchSteps.length - 1)) * 100}%`,
            height: "100%",
            backgroundColor: "#1a7dc5",
            transition: "width 0.3s ease",
          }}></div>
      </div>
      <div
        className="controls"
        style={{
          display: "flex",
          justifyContent: "center",
          gap: "10px",
          marginBottom: "20px",
        }}>
        <button onClick={() => setCurrentStep(0)} style={{ padding: "5px 10px" }}>
          ⏮️ First
        </button>

        <button
          onClick={() => setCurrentStep((prev) => Math.max(0, prev - 1))}
          disabled={currentStep === 0}
          style={{ padding: "5px 10px" }}>
          ⏪ Prev
        </button>

        <button onClick={() => setIsPlaying(!isPlaying)} style={{ padding: "5px 10px" }}>
          {isPlaying ? "⏸️ Pause" : "▶️ Play"}
        </button>

        <button
          onClick={() => setCurrentStep((prev) => Math.min(searchSteps.length - 1, prev + 1))}
          disabled={currentStep === searchSteps.length - 1}
          style={{ padding: "5px 10px" }}>
          ⏩ Next
        </button>

        <button onClick={() => setCurrentStep(searchSteps.length - 1)} style={{ padding: "5px 10px" }}>
          ⏭️ Last
        </button>

        <div style={{ display: "flex", alignItems: "center" }}>
          <label style={{ marginRight: "8px" }}>Speed:</label>
          <input
            type="range"
            min="100"
            max="2000"
            step="100"
            value={playbackSpeed}
            onChange={(e) => setPlaybackSpeed(Number(e.target.value))}
          />
        </div>
      </div>
      <div className="visualization-container">
        <SearchTreeView
          discovered={discovered}
          currentNode={step.current}
          queueNodes={step.queue || []}
          seenNodes={step.seen || []}
        />
      </div>
    </div>
  );
}

export default LiveSearchVisualizer;
