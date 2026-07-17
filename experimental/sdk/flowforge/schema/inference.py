"""Schema inference from Python type hints."""

from __future__ import annotations
from typing import Any, Dict, get_origin, get_args, Union, List, Optional, Type
import json


def infer_schema(type_hint: Optional[Type]) -> Dict[str, Any]:
    """Infer JSON Schema from Python type hint.
    
    Args:
        type_hint: Python type hint
    
    Returns:
        JSON Schema as dictionary
    
    Examples:
        >>> infer_schema(int)
        {'type': 'integer'}
        
        >>> infer_schema(list)
        {'type': 'array'}
        
        >>> infer_schema(List[Dict[str, int]])
        {'type': 'array', 'items': {'type': 'object', 'properties': {...}}}
    """
    if type_hint is None:
        return {"type": "object"}
    
    # Handle Optional[X] -> Union[X, None]
    if get_origin(type_hint) is Union:
        args = get_args(type_hint)
        if type(None) in args:
            # Optional[X]
            non_none_args = [arg for arg in args if arg is not type(None)]
            if len(non_none_args) == 1:
                schema = infer_schema(non_none_args[0])
                # Add null to allowed types
                if isinstance(schema.get('type'), str):
                    schema['type'] = [schema['type'], 'null']
                return schema
        else:
            # Union[X, Y, ...] - use first non-None type
            return infer_schema(args[0])
    
    # Handle List[X]
    if get_origin(type_hint) is list:
        args = get_args(type_hint)
        schema = {"type": "array"}
        if args:
            schema["items"] = infer_schema(args[0])
        return schema
    
    # Handle Dict[K, V]
    if get_origin(type_hint) is dict:
        return {"type": "object"}
    
    # Handle basic types
    if type_hint is int:
        return {"type": "integer"}
    elif type_hint is str:
        return {"type": "string"}
    elif type_hint is float:
        return {"type": "number"}
    elif type_hint is bool:
        return {"type": "boolean"}
    elif type_hint is list:
        return {"type": "array"}
    elif type_hint is dict:
        return {"type": "object"}
    
    # Default to object
    return {"type": "object"}


def validate_data(data: Any, schema: Dict[str, Any]) -> tuple[bool, List[str]]:
    """Validate data against JSON Schema.
    
    Args:
        data: Data to validate
        schema: JSON Schema
    
    Returns:
        (is_valid, error_messages)
    
    Note:
        This is a simplified validator. For complex schemas, use jsonschema library.
    """
    errors = []
    schema_type = schema.get('type')
    
    # Check type
    if schema_type == 'integer':
        if not isinstance(data, int) or isinstance(data, bool):
            errors.append(f"Expected integer, got {type(data).__name__}")
    
    elif schema_type == 'string':
        if not isinstance(data, str):
            errors.append(f"Expected string, got {type(data).__name__}")
    
    elif schema_type == 'number':
        if not isinstance(data, (int, float)) or isinstance(data, bool):
            errors.append(f"Expected number, got {type(data).__name__}")
    
    elif schema_type == 'boolean':
        if not isinstance(data, bool):
            errors.append(f"Expected boolean, got {type(data).__name__}")
    
    elif schema_type == 'array':
        if not isinstance(data, list):
            errors.append(f"Expected array, got {type(data).__name__}")
        elif 'items' in schema:
            item_schema = schema['items']
            for i, item in enumerate(data):
                valid, item_errors = validate_data(item, item_schema)
                errors.extend([f"[{i}] {e}" for e in item_errors])
    
    elif schema_type == 'object':
        if not isinstance(data, dict):
            errors.append(f"Expected object, got {type(data).__name__}")
    
    return len(errors) == 0, errors


__all__ = ['infer_schema', 'validate_data']
