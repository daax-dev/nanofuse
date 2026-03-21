---
description: Execute research and business validation workflow using specialized agents.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

This command orchestrates research and business validation using two specialized agents:

1. **Researcher Agent**: Goes out and does research on an idea
   - Conducts market research
   - Analyzes competitive landscape
   - Gathers technical feasibility data
   - Identifies industry trends and best practices

2. **Business Validator Agent**: Provides realistic validation based on certain criteria
   - Evaluates business viability
   - Assesses market potential
   - Validates against business constraints
   - Provides risk assessment

## Execution Flow

1. Parse user input to understand research topic and validation requirements

2. **Research Phase**:
   - Dispatch Researcher agent to gather information
   - Collect findings on market, technical, and competitive aspects
   - Document research results

3. **Validation Phase**:
   - Dispatch Business Validator agent with research findings
   - Apply validation criteria (market size, feasibility, ROI potential, etc.)
   - Generate validation report

4. **Output**:
   - Consolidated research document
   - Business validation assessment
   - Recommendations for next steps

## Notes

- This command is a placeholder for future agent implementation
- Full agent integration will be completed in a future task
