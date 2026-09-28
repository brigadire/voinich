# Generator semantics

For transformed source stream S and target T, the generator enumerates partial injective functions from source-unit types to EVA-unit types. A source unit is either globally unmapped or mapped on its first occurrence; all later occurrences obey the same decision. The output is the full global replay of S with unmapped units removed. Required mappings are the mapped pairs; forbidden source units are all source types chosen unmapped.
