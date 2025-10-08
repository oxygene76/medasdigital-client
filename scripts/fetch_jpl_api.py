#!/usr/bin/env python3
"""
Fetch solar system data from JPL Horizons using direct API calls
No external dependencies required - uses only standard library
"""

import json
import urllib.request
import urllib.parse
from datetime import datetime
import re

class JPLHorizonsAPI:
    """Direct API interface to JPL Horizons"""
    
    def __init__(self):
        self.base_url = "https://ssd.jpl.nasa.gov/api/horizons.api"
        
    def fetch_elements(self, object_id, epoch="2023-09-13"):
        """
        Fetch orbital elements for an object
        
        Parameters:
        - object_id: JPL identifier (e.g., '599' for Jupiter, '90377' for Sedna)
        - epoch: Date string for epoch
        """
        
        # Build query parameters
        params = {
            'format': 'text',
            'COMMAND': f"'{object_id}'",
            'OBJ_DATA': 'YES',
            'MAKE_EPHEM': 'YES',
            'EPHEM_TYPE': 'ELEMENTS',
            'CENTER': "'500@10'",  # Sun
            'START_TIME': f"'{epoch}'",
            'STOP_TIME': f"'{epoch}'",
            'STEP_SIZE': '1 d',
            'REF_SYSTEM': 'ICRF',
            'REF_PLANE': 'ECLIPTIC',
            'OUT_UNITS': 'AU-D',
            'CSV_FORMAT': 'NO'
        }
        
        # Build URL
        url = self.base_url + '?' + urllib.parse.urlencode(params)
        
        try:
            # Make request
            with urllib.request.urlopen(url) as response:
                data = response.read().decode('utf-8')
                
            # Parse the response
            return self.parse_horizons_output(data, object_id)
            
        except Exception as e:
            print(f"    Error fetching {object_id}: {e}")
            return None
    
    def parse_horizons_output(self, text, object_id):
        """Parse JPL Horizons text output for orbital elements"""
        
        elements = {}
        
        # Check if object was found
        if "No matches found" in text or "Cannot find" in text:
            return None
            
        # Parse orbital elements section
        if "$$SOE" in text and "$$EOE" in text:
            # Extract data between markers
            start = text.find("$$SOE") + 5
            end = text.find("$$EOE")
            data_section = text[start:end].strip()
            
            # Parse the elements line by line
            lines = data_section.split('\n')
            
            # Elements are typically in this format
            # EC, QR, IN, OM, W, Tp, N, MA, TA, A, AD, PR
            for line in lines:
                # Look for lines with orbital elements (contain numbers and commas)
                if ',' in line and any(c.isdigit() for c in line):
                    parts = [p.strip() for p in line.split(',')]
                    if len(parts) >= 11:
                        try:
                            # Standard orbital elements from JPL
                            elements = {
                                "eccentricity": float(parts[0]) if parts[0] else 0,
                                "perihelion": float(parts[1]) if parts[1] else 0,
                                "inclination": float(parts[2]) if parts[2] else 0,
                                "longitude_ascending_node": float(parts[3]) if parts[3] else 0,
                                "argument_perihelion": float(parts[4]) if parts[4] else 0,
                                "mean_anomaly": float(parts[7]) if parts[7] else 0,
                                "semimajor_axis": float(parts[9]) if parts[9] else 0,
                                "aphelion": float(parts[10]) if parts[10] else 0,
                                "orbital_period": float(parts[11]) if parts[11] else None
                            }
                        except (ValueError, IndexError):
                            pass
        
        # Alternative: Parse from object data section
        if not elements:
            # Look for individual element patterns in the text
            patterns = {
                "semimajor_axis": r"a\s*=\s*([\d.+-]+)",
                "eccentricity": r"e\s*=\s*([\d.+-]+)",
                "inclination": r"i\s*=\s*([\d.+-]+)",
                "longitude_ascending_node": r"node\s*=\s*([\d.+-]+)|Omega\s*=\s*([\d.+-]+)",
                "argument_perihelion": r"peri\s*=\s*([\d.+-]+)|w\s*=\s*([\d.+-]+)",
                "mean_anomaly": r"M\s*=\s*([\d.+-]+)",
                "perihelion": r"q\s*=\s*([\d.+-]+)",
                "aphelion": r"Q\s*=\s*([\d.+-]+)"
            }
            
            for key, pattern in patterns.items():
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    value = match.group(1) if match.group(1) else match.group(2)
                    try:
                        elements[key] = float(value)
                    except ValueError:
                        pass
        
        return elements if elements else None

def get_planet_data():
    """Get planet data with known masses"""
    return [
        {"name": "Jupiter", "id": "599", "mass": 0.0009545942},
        {"name": "Saturn", "id": "699", "mass": 0.0002857214},
        {"name": "Uranus", "id": "799", "mass": 0.00004365785},
        {"name": "Neptune", "id": "899", "mass": 0.00005149497}
    ]

def get_tno_list():
    """Get list of TNOs to fetch"""
    return [
        # Well-established ETNOs
        {"name": "Sedna", "id": "90377", "designation": "2003 VB12"},
        {"name": "Eris", "id": "136199", "designation": "2003 UB313"},
        {"name": "Makemake", "id": "136472", "designation": "2005 FY9"},
        {"name": "Gonggong", "id": "225088", "designation": "2007 OR10"},
        
        # Recent discoveries with extreme orbits
        {"name": "2012 VP113", "id": "2012 VP113", "designation": "2012 VP113"},
        {"name": "The Goblin", "id": "2015 TG387", "designation": "2015 TG387"},
        {"name": "2004 VN112", "id": "2004 VN112", "designation": "2004 VN112"},
        {"name": "2010 GB174", "id": "2010 GB174", "designation": "2010 GB174"},
        {"name": "2013 SY99", "id": "2013 SY99", "designation": "2013 SY99"},
        {"name": "2014 SR349", "id": "2014 SR349", "designation": "2014 SR349"},
        {"name": "2014 FE72", "id": "2014 FE72", "designation": "2014 FE72"},
        
        # Additional significant TNOs
        {"name": "Quaoar", "id": "50000", "designation": "2002 LM60"},
        {"name": "Orcus", "id": "90482", "designation": "2004 DW"},
        {"name": "Haumea", "id": "136108", "designation": "2003 EL61"},
        {"name": "Varuna", "id": "20000", "designation": "2000 WR106"},
    ]

def main():
    """Main execution"""
    
    print("=" * 60)
    print("JPL Horizons Data Fetcher (Direct API)")
    print("=" * 60)
    
    api = JPLHorizonsAPI()
    epoch = "2023-09-13"
    
    # Fetch planet data
    print("\nFetching Planet Data...")
    print("-" * 40)
    planets = []
    for planet_info in get_planet_data():
        print(f"  {planet_info['name']}...", end=" ")
        elements = api.fetch_elements(planet_info['id'], epoch)
        if elements:
            planets.append({
                "name": planet_info['name'],
                "id": planet_info['id'],
                "mass": planet_info['mass'],
                "orbital_elements": elements
            })
            print("✓")
        else:
            print("✗")
    
    # Fetch TNO data
    print("\nFetching TNO/ETNO Data...")
    print("-" * 40)
    tnos = []
    for tno_info in get_tno_list():
        print(f"  {tno_info['name']:20s}...", end=" ")
        
        # Try different ID formats
        elements = None
        for id_format in [tno_info['id'], f"DES={tno_info['id']}"]:
            elements = api.fetch_elements(id_format, epoch)
            if elements:
                break
        
        if elements:
            tno_type = "extreme_tno" if elements.get('perihelion', 0) > 30 else "tno"
            tnos.append({
                "name": tno_info['name'],
                "designation": tno_info['designation'],
                "type": tno_type,
                "orbital_elements": elements
            })
            print("✓")
        else:
            print("✗")
    
    # Create output structure
    output = {
        "metadata": {
            "version": datetime.now().strftime("%Y.%m.%d"),
            "source": "JPL Horizons System",
            "url": "https://ssd.jpl.nasa.gov/horizons/",
            "fetch_date": datetime.now().isoformat(),
            "epoch_jd": 2460200.5,
            "epoch_date": epoch + " 00:00:00 UTC",
            "units": {
                "distance": "AU",
                "angle": "degrees",
                "mass": "solar_masses",
                "time": "days"
            }
        },
        "planets": planets,
        "etnos": tnos,
        "planet9_search_parameters": {
            "mass_range": [5, 10],
            "mass_unit": "earth_masses",
            "semimajor_axis_range": [400, 800],
            "eccentricity_range": [0.2, 0.5],
            "inclination_range": [15, 30],
            "argument_perihelion_clustering": {
                "center": 318,
                "spread": 30,
                "note": "Based on Batygin & Brown 2016 clustering analysis"
            },
            "references": [
                "Batygin, K. & Brown, M. E. (2016). Evidence for a Distant Giant Planet in the Solar System. AJ 151, 22.",
                "Trujillo, C. A. & Sheppard, S. S. (2014). A Sedna-like body with a perihelion of 80 AU. Nature 507, 471.",
                "Sheppard, S. S. et al. (2019). A New High Perihelion Trans-Plutonian Inner Oort Cloud Object. AJ 157, 139."
            ]
        }
    }
    
    # Save to file
    output_file = "data/solar_system_jpl.json"
    with open(output_file, 'w') as f:
        json.dump(output, f, indent=2)
    
    # Print summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Successfully fetched:")
    print(f"  • {len(planets)} planets")
    print(f"  • {len(tnos)} TNOs/ETNOs")
    print(f"\nData saved to: {output_file}")
    
    # Clustering analysis
    if tnos:
        print("\n" + "=" * 60)
        print("ETNO CLUSTERING ANALYSIS")
        print("=" * 60)
        
        extreme_tnos = [t for t in tnos if t.get('type') == 'extreme_tno']
        if extreme_tnos:
            print(f"Found {len(extreme_tnos)} extreme TNOs (q > 30 AU):\n")
            
            for tno in extreme_tnos:
                elem = tno['orbital_elements']
                lon_peri = (elem.get('longitude_ascending_node', 0) + 
                           elem.get('argument_perihelion', 0)) % 360
                print(f"  {tno['name']:20s}: ϖ = {lon_peri:6.1f}° "
                      f"(q={elem.get('perihelion', 0):.1f} AU)")

if __name__ == "__main__":
    main()
