---
description: Execute planning workflow using project architect and platform engineer agents (builds out /speckit.constitution).
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

This command executes the planning workflow using two specialized agents:

1. **Project Architect Agent**: Designs architecture like Gregor Hohpe
   - Creates comprehensive system architecture
   - Defines architectural patterns and principles
   - Designs component interactions and boundaries
   - Establishes architectural decision records (ADRs)
   - Builds out /speckit.constitution with architectural principles

2. **Platform Engineer Agent**: Designs infrastructure and backend systems
   - Plans infrastructure architecture
   - Designs backend systems and services
   - Defines deployment and scaling strategies
   - Establishes platform engineering best practices
   - Contributes platform principles to /speckit.constitution

## Execution Flow

1. Parse user input to understand planning requirements and constraints

2. **Architecture Planning**:
   - Dispatch Project Architect agent
   - Generate system architecture design
   - Create architectural diagrams and documentation
   - Define architectural principles and patterns
   - Document key architectural decisions

3. **Platform Planning**:
   - Dispatch Platform Engineer agent
   - Design infrastructure architecture
   - Plan backend systems and data flows
   - Define deployment pipelines and strategies
   - Establish observability and monitoring approach

4. **Constitution Integration**:
   - Build out /speckit.constitution with:
     - Architectural principles and constraints
     - Platform engineering standards
     - Infrastructure requirements
     - Deployment and operational guidelines

5. **Output**:
   - Complete implementation plan
   - Architecture design documents
   - Infrastructure design specifications
   - Updated constitution with planning principles
   - Technical decision records

## Notes

- This command is a placeholder for future agent implementation
- Full Project Architect and Platform Engineer agent integration will be completed in a future task
- Builds out and updates /speckit.constitution
